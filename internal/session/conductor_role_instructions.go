package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const ConductorRoleInstructionsFileName = "ROLE_INSTRUCTIONS.md"

func isFileBackedConductor(inst *Instance) bool {
	return conductorNameFromInstance(inst) != ""
}

func ConductorRoleInstructionsPath(name string) (string, error) {
	dir, err := ConductorNameDir(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConductorRoleInstructionsFileName), nil
}

func ConductorRoleInstructionsPathForInstance(inst *Instance) (string, error) {
	name := conductorNameFromInstance(inst)
	if name == "" {
		return "", fmt.Errorf("instance %q is not a named conductor", inst.Title)
	}
	return ConductorRoleInstructionsPath(name)
}

func ReadConductorRoleInstructions(name string) (string, error) {
	path, err := ConductorRoleInstructionsPath(name)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read conductor role instructions: %w", err)
	}
	return strings.TrimRight(string(data), "\n"), nil
}

func ReadPersistentRoleInstructions(inst *Instance) (string, error) {
	if !isFileBackedConductor(inst) {
		return inst.RoleInstructions, nil
	}
	content, err := ReadConductorRoleInstructions(conductorNameFromInstance(inst))
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(content) != "" {
		return content, nil
	}
	return inst.RoleInstructions, nil
}

func WriteConductorRoleInstructions(name, value string) error {
	path, err := ConductorRoleInstructionsPath(name)
	if err != nil {
		return err
	}
	trimmed := strings.TrimRight(value, "\n")
	if strings.TrimSpace(trimmed) == "" {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove conductor role instructions: %w", err)
		}
		return nil
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create conductor role instructions dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(trimmed), 0o644); err != nil {
		return fmt.Errorf("write conductor role instructions: %w", err)
	}
	return nil
}

func SavePersistentRoleInstructions(inst *Instance, value string) error {
	value = strings.TrimRight(value, "\n")
	if isFileBackedConductor(inst) {
		if err := WriteConductorRoleInstructions(conductorNameFromInstance(inst), value); err != nil {
			return err
		}
	}
	inst.RoleInstructions = value
	return nil
}

func backfillConductorRoleInstructions(inst *Instance) error {
	if !isFileBackedConductor(inst) || strings.TrimSpace(inst.RoleInstructions) == "" {
		return nil
	}
	name := conductorNameFromInstance(inst)
	existing, err := ReadConductorRoleInstructions(name)
	if err != nil {
		return err
	}
	if strings.TrimSpace(existing) != "" {
		return nil
	}
	return WriteConductorRoleInstructions(name, inst.RoleInstructions)
}

func PrepareConductorRoleInstructionsForStart(inst *Instance) error {
	if !isFileBackedConductor(inst) {
		return nil
	}
	name := conductorNameFromInstance(inst)
	meta, err := LoadConductorMeta(name)
	if err == nil {
		if err := RefreshConductorInstructionFilesForRoleOverlay(name, meta.Profile, meta.GetAgent()); err != nil {
			return err
		}
	}
	return backfillConductorRoleInstructions(inst)
}
