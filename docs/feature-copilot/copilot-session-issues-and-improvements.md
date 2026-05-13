# Copilot セッション: 課題と改善メモ

> 出典: agent-deck で 5 つの並列 Copilot セッションを運用した実測メモ（bid-sys プロジェクト）  
> 日付: 2026-05-02

---

## 1. セッション作成まわりの課題

### 1.1 `-c` 引数の渡し方で挙動が変わる

`-c` の指定方法によって、agent-deck が認識する `tool` 種別が変わり、動作も大きく変わる。

| 指定例 | 認識される tool | `-g <group>` 利用可否 | 補足 |
|--------|------------------|-----------------------|------|
| `-c copilot` | `tool:copilot` | ❌ `"path does not exist"` になる | group 名がパスとして解釈される |
| `-c "copilot --allow-all --model claude-opus-4.6"` | `tool:shell` | ✅（path から自動推定） | スペースを含む文字列は shell 経由になる |

**原因**: コマンド文字列にスペースがあると agent-deck は shell コマンドとして扱い、単語 1 個なら `tool:copilot` の専用経路に入る。両者で group 解決ロジックが異なる。

**現時点で使える書き方**:

```bash
agent-deck add <worktree_path> \
  -t "my-session" \
  -c "copilot --allow-all --model claude-opus-4.6"
# -g は渡さず、path から group を自動推定させる
```

---

### 1.2 `-g <group>` の解決失敗

**現象**: conductor セッション（cwd = `~/.agent-deck/conductor/ops`）から次を実行すると失敗する。

```bash
agent-deck add /path/to/worktree -t "title" -c copilot -g "bid-sys"
# Error: path does not exist
```

**原因**: agent-deck は `-g` の値を group 識別子ではなく、現在の cwd から解決する**パス**として扱う。conductor の cwd から `bid-sys` は見つからないためエラーになる。

**回避策**: `-g` を付けず、worktree の絶対パスを `<path>` に渡す。agent-deck が `.worktrees/` の親ディレクトリ名などから group を自動推定する。

**改善案**: `-g` はまず既存 group 名として解決し、失敗した場合だけパスとして扱う。あるいは `--group-name` と `--group-path` を分ける。

---

### 1.3 `-extra-arg` は Copilot では効かない

**現象**: `-extra-arg` で `--allow-all --model claude-opus-4.6` を Copilot に渡しても反映されない。

**原因**: `-extra-arg` は `-c claude` 系の経路にしか適用されず、他の tool には効かない。

**回避策**: 引数を `-c` に直接埋め込む。

```bash
-c "copilot --allow-all --model claude-opus-4.6"
```

**改善案**: `-extra-arg` 相当を全 tool で使えるようにするか、`[tools.copilot]` のような per-tool 設定を用意する。

---

### 1.4 `agent-deck launch` と `agent-deck add` の差

**現象**: conductor の cwd（`~/.agent-deck/conductor/ops`）から `agent-deck launch <path> -g "bid-sys"` を実行すると、上記の group 解決問題にぶつかる。

**回避策**: いったん `add` してから `session start` する。

```bash
agent-deck add <worktree_path> -t "title" -c "copilot --allow-all --model claude-opus-4.6"
agent-deck session start <title>
# 通信は tmux send-keys を使う。session send は使えない
```

---

## 2. セッション通信まわりの課題

### 2.1 `agent-deck session send` が Copilot セッションで効かない

**現象**:

```bash
agent-deck session send my-copilot-session "do something"
# コマンド自体は成功するが、Copilot 側には届かない
```

**原因**: Copilot CLI には ACP（Agent Communication Protocol）統合がなく、当時の agent-deck でも Claude / Codex / Gemini のような hook ベースの lifecycle tracking（activity 検出、session-id 取得、`--resume` など）が未実装だった。

**回避策**: `tmux send-keys` で Copilot の TUI に直接キー入力する。

```bash
tmux send-keys -t <tmux_session_name> "your message here" Enter
```

**追記 (2026-05-13)**: 最新の agent-deck では Copilot CLI hooks を使った lifecycle tracking が入り、`Stop` などの完了系イベントは status/transition 検知に流れるようになった。`session send` 自体の stdin/ACP 問題は別件として残る。

---

### 2.2 `tmux send-keys` の Enter が Copilot TUI で送信扱いされない

**現象**:

```bash
tmux send-keys -t sess "Read AGENTS.md and implement all tasks" Enter
# 入力欄には文字が出るが、Enter で送信されない
```

**原因**: Copilot TUI は readline 系のインターフェースを使っており、tmux control mode 由来の仮想 `Enter` キーイベントを、実キーボード入力と同じようには扱わない。

**確認済みの回避策**（信頼度順）:

1. `Enter` の代わりに `\r` を送る

```bash
tmux send-keys -t sess "your message" $'\r'
# または
tmux send-keys -t sess "your message"$'\r'
```

2. 先に `C-c` で入力欄をクリアし、1 文字ずつ送って最後に `\r`

```bash
tmux send-keys -t sess C-c
sleep 0.2
echo "your message" | while IFS= read -rn1 char; do
  tmux send-keys -t sess "$char"
  sleep 0.02
done
tmux send-keys -t sess $'\r'
```

3. Copilot CLI が受け付けるなら `/task` プレフィックスを使う

```bash
tmux send-keys -t sess "/task Read AGENTS.md and implement all tasks" $'\r'
```

4. 非対話モードに切り替える（最も確実だが、継続対話は失う）

```bash
copilot -p "Read AGENTS.md and implement all tasks"
```

---

## 3. 初回起動時の課題

### 3.1 新しいディレクトリでのアクセス許可ダイアログ

**現象**: Copilot を新しいディレクトリで初回起動すると、対話式の許可ダイアログが出る。

```text
Do you want to allow access to this directory?
❯ 1. Yes
  2. Yes, and don't ask again for <tool> in this directory
  3. No
```

スクリプト実行中にこのダイアログが出ると、後続処理がすべて止まる。

**対処方法**:

```bash
# 起動後、ダイアログ表示を待って "2" を送る
sleep 3
tmux send-keys -t sess "2" Enter
```

---

## 4. モデル指定まわりの課題

### 4.1 モデル名の形式

**有効な形式**（status bar では "Claude Opus 4.6" と表示される）:

```bash
-c "copilot --allow-all --model claude-opus-4.6"
```

**無効だった形式**: `claude-opus-4.5`（当時は利用枠やサポート対象外でエラー）

**改善案**: 現在は tmux status bar を目視しないと model を確認できないので、session 詳細に `model` を露出するとスクリプトから検証しやすい。

---

## 5. 改善提案まとめ

| 優先度 | 課題 | 提案 |
|--------|------|------|
| P0 | `session send` が Copilot で使えない | Copilot hook 統合を実装する（stdin か ACP） |
| P0 | `tmux send-keys` の Enter で送信されない | 文書化して `\r` を案内する、または内部的に `\r` を使う |
| P1 | `-g` の解釈が曖昧 | group 名と group パスを分離するか、先に名前で解決する |
| P1 | `-extra-arg` が Claude 限定 | 全 tool に広げるか、per-tool config を導入する |
| P2 | 初回許可ダイアログを自動化しづらい | `--allow-all` 時の自動許可や `--no-interactive` を用意する |
| P2 | model を問い合わせられない | `session show --json` に `model` を含める |

---

## 6. 参考

- [issue #556: Add support for GitHub Copilot CLI](https://github.com/asheshgoplani/agent-deck/issues/556)
- [GitHub Copilot CLI best practices: configure allowed tools](https://docs.github.com/en/copilot/how-tos/copilot-cli/cli-best-practices#configure-allowed-tools)
