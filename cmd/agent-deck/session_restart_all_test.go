package main

import (
	"strings"
	"testing"
	"time"

	"github.com/asheshgoplani/agent-deck/internal/session"
)

func TestSelectRestartAllTargets_ClassifiesAliveRecoverableAndSkipped(t *testing.T) {
	alive := session.NewInstance("alive", "/tmp")
	alive.Status = session.StatusRunning

	recoverableWaiting := session.NewInstance("recoverable-waiting", "/tmp")
	recoverableWaiting.Status = session.StatusWaiting

	recoverableIdle := session.NewInstance("recoverable-idle", "/tmp")
	recoverableIdle.Status = session.StatusIdle
	recoverableIdle.LastStartedAt = time.Now().Add(-2 * time.Minute)

	stopped := session.NewInstance("stopped", "/tmp")
	stopped.Status = session.StatusStopped

	neverStarted := session.NewInstance("never-started", "/tmp")
	neverStarted.Status = session.StatusIdle

	targets, skipped := selectRestartAllTargets(
		[]*session.Instance{alive, recoverableWaiting, recoverableIdle, stopped, neverStarted},
		func(inst *session.Instance) bool { return inst.ID == alive.ID },
	)

	if len(targets) != 3 {
		t.Fatalf("expected 3 restart targets, got %d", len(targets))
	}

	if targets[0].Instance.ID != alive.ID || targets[0].Action != restartAllActionRestart {
		t.Fatalf("target[0] = %+v, want alive restart", targets[0])
	}
	if targets[1].Instance.ID != recoverableWaiting.ID || targets[1].Action != restartAllActionRecover {
		t.Fatalf("target[1] = %+v, want waiting recovery", targets[1])
	}
	if targets[2].Instance.ID != recoverableIdle.ID || targets[2].Action != restartAllActionRecover {
		t.Fatalf("target[2] = %+v, want idle-with-history recovery", targets[2])
	}

	if len(skipped) != 2 {
		t.Fatalf("expected 2 skipped results, got %d", len(skipped))
	}
	if skipped[0].Title != stopped.Title || skipped[0].Reason != restartAllSkipReasonStopped {
		t.Fatalf("skipped[0] = %+v, want stopped skip", skipped[0])
	}
	if skipped[1].Title != neverStarted.Title || skipped[1].Reason != restartAllSkipReasonNeverStarted {
		t.Fatalf("skipped[1] = %+v, want never-started skip", skipped[1])
	}
}

func TestSummarizeRestartAllResults_TracksRecoveredSkippedAndFailed(t *testing.T) {
	results := []restartAllResult{
		{Title: "alive", Action: string(restartAllActionRestart), Success: true},
		{Title: "dead-ok", Action: string(restartAllActionRecover), Success: true},
		{Title: "dead-fail", Action: string(restartAllActionRecover), Success: false, Error: "boom"},
		{Title: "stopped", Action: string(restartAllActionSkipped), Success: true, Skipped: true, Reason: restartAllSkipReasonStopped},
	}

	summary := summarizeRestartAllResults(results)

	if summary.Total != 4 {
		t.Fatalf("Total = %d, want 4", summary.Total)
	}
	if summary.Restarted != 1 {
		t.Fatalf("Restarted = %d, want 1", summary.Restarted)
	}
	if summary.Recovered != 1 {
		t.Fatalf("Recovered = %d, want 1", summary.Recovered)
	}
	if summary.Skipped != 1 {
		t.Fatalf("Skipped = %d, want 1", summary.Skipped)
	}
	if summary.Failed != 1 {
		t.Fatalf("Failed = %d, want 1", summary.Failed)
	}

	line := summary.Format()
	for _, sub := range []string{"restarted=1", "recovered=1", "skipped=1", "failed=1", "total=4"} {
		if !strings.Contains(line, sub) {
			t.Fatalf("summary format %q missing %q", line, sub)
		}
	}
}
