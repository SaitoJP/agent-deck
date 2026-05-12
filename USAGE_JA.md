# Agent Deck 使い方ガイド

Agent Deck は、**tmux 上で AI エージェントのセッションをまとめて管理するための CLI / TUI** です。  
Claude Code、GitHub Copilot CLI、Codex などを並列で起動し、状態確認、切り替え、監督、再開を 1 つの場所から行えます。

このガイドは、**Claude / Copilot / Codex のどれを使っても分かる**ことを重視しつつ、最後に **SSH サーバ上で Conductor と子セッションを継続運用する手順** までまとめています。

## 1. まず押さえる用語

- **session**: 1 本の AI 作業セッション
- **conductor**: 他の session を監督する長寿命セッション
- **group**: セッションのまとまり
- **tmux**: SSH 切断後もセッションを残すための土台

つまり Agent Deck は、**複数の AI セッションを tmux で持続させながら管理する司令塔**です。

## 2. 前提

最低限、以下を確認してください。

```bash
tmux -V
agent-deck --help
```

そのうえで、使いたいランタイムがサーバ上で利用可能である必要があります。

| 使うもの | 確認コマンド |
|---|---|
| Claude Code | `claude --help` |
| GitHub Copilot CLI | `copilot --help` |
| Codex | `codex --help` |

## 3. どのランタイムを使うか

Agent Deck では、子セッションと Conductor の両方にランタイムを選べます。

| 用途 | Claude | Copilot | Codex |
|---|---|---|---|
| 子セッション作成 | `-c claude` | `-c "copilot --allow-all --model claude-sonnet-4.6"` | `-c codex` |
| Conductor 作成 | 既定値のままで可 | `--agent copilot` | `--agent codex` |

補足:

- **Claude / Copilot の conductor** は `CLAUDE.md` を使います
- **Codex の conductor** は `AGENTS.md` を使います
- Copilot は `--model` や `--allow-all` を付ける運用が分かりやすいです

## 4. 一番基本的な使い方

まずは通常のセッションを 1 本作る流れです。

パスには **`~/work/myrepo` のような明示パス** も **`.` のようなカレントディレクトリ** も使えます。

- **いまいるリポジトリを対象**にするなら `.` で OK
- **別のリポジトリや worktree を対象**にするなら明示パスが安全
- 特に conductor 配下や別ディレクトリから操作するときは、対象を誤解しにくいよう **明示パス推奨**です

### Claude の例

```bash
agent-deck add ~/work/myrepo -t my-task -c claude
agent-deck session start my-task
agent-deck list
agent-deck session output my-task
```

同じ意味で、次のように `.` でも書けます。

```bash
agent-deck add . -t my-task -c claude
agent-deck launch . -t my-task -c codex -m "このディレクトリを調べて概要を要約してください。"
```

### Copilot の例

```bash
agent-deck add ~/work/myrepo \
  -t my-task \
  -c "copilot --allow-all --model claude-sonnet-4.6"

agent-deck session start my-task
agent-deck list
agent-deck session output my-task
```

### Codex の例

```bash
agent-deck add ~/work/myrepo -t my-task -c codex
agent-deck session start my-task
agent-deck list
agent-deck session output my-task
```

よく使うコマンドはこの 4 つです。

| コマンド | 役割 |
|---|---|
| `agent-deck add ...` | セッションを作る |
| `agent-deck launch ...` | セッション作成と起動を一度に行う |
| `agent-deck list` | 全セッションの状態を見る |
| `agent-deck session output <name>` | そのセッションの出力を見る |

## 5. Conductor の考え方

Conductor は、あなたの代わりに子セッションを見張る**監督役**です。

たとえば次のように使います。

- `backend` が調査担当
- `frontend` が UI 担当
- `conductor-ops` が全体監督

Conductor 自体も 1 本の session として管理されます。  
`agent-deck conductor setup ops` を実行すると、実際の session 名は **`conductor-ops`** になります。

## 6. Conductor を作る

### Claude の例

```bash
agent-deck conductor setup ops \
  --description "Claude conductor for coding"

agent-deck session start conductor-ops
```

### Copilot の例

```bash
agent-deck conductor setup ops \
  --agent copilot \
  --model claude-sonnet-4.6 \
  --allow-all \
  --description "Copilot conductor for coding"

agent-deck session start conductor-ops
```

### Codex の例

```bash
agent-deck conductor setup ops \
  --agent codex \
  --description "Codex conductor for coding"

agent-deck session start conductor-ops
```

確認:

```bash
agent-deck conductor status ops
agent-deck session output conductor-ops
```

## 7. 子セッションを作る

### 方法 A: `add` → `start`

#### Claude

```bash
agent-deck add ~/work/myrepo -t api-fix -c claude
agent-deck session start api-fix
```

#### Copilot

```bash
agent-deck add ~/work/myrepo \
  -t api-fix \
  -c "copilot --allow-all --model claude-sonnet-4.6"

agent-deck session start api-fix
```

#### Codex

```bash
agent-deck add ~/work/myrepo -t api-fix -c codex
agent-deck session start api-fix
```

### 方法 B: `launch` で一発起動

#### Claude

```bash
agent-deck launch ~/work/myrepo \
  -t ui-task \
  -c claude \
  -m "このリポジトリを開き、コードベースの構成を調べて、概要を要約してください。"
```

#### Copilot

```bash
agent-deck launch ~/work/myrepo \
  -t ui-task \
  -c "copilot --allow-all --model claude-sonnet-4.6" \
  -m "このリポジトリを開き、コードベースの構成を調べて、概要を要約してください。"
```

#### Codex

```bash
agent-deck launch ~/work/myrepo \
  -t ui-task \
  -c codex \
  -m "このリポジトリを開き、コードベースの構成を調べて、概要を要約してください。"
```

確認:

```bash
agent-deck list
agent-deck session output api-fix
agent-deck session output ui-task
```

## 8. Conductor に監督させる

Conductor への指示は、どのランタイムでも基本的に同じです。

```bash
agent-deck session send conductor-ops "あなたはコーディング用セッションの監督役です。api-fix と ui-task を監視し、waiting や error の状態を確認し、それぞれが今何をしているかを報告してください。"
```

確認:

```bash
agent-deck session output conductor-ops
```

## 9. 子セッションに作業を依頼する

まずは標準の送り方です。

```bash
agent-deck session send api-fix "このリポジトリで発生している不具合を調査し、修正方針を提案してください。"
agent-deck session send ui-task "UI のエントリポイントを特定し、主要コンポーネントを把握したうえで、必要な変更の実装を開始してください。"
```

通常はこれで進められます。  
ただし **Copilot セッションでは `session send` が環境によって届かないことがあります。** その場合は tmux に直接入力します。

```bash
tmux send-keys -t api-fix C-c
tmux send-keys -t api-fix "このリポジトリで発生している不具合を調査し、修正方針を提案してください。" $'\r'

tmux send-keys -t ui-task C-c
tmux send-keys -t ui-task "UI のエントリポイントを特定し、主要コンポーネントを把握したうえで、必要な変更の実装を開始してください。" $'\r'
```

Copilot では `Enter` より **`$'\r'`** の方が通りやすいことがあります。

## 10. SSH を切っても継続させる

Agent Deck は tmux 管理下で動くため、適切に起動していれば SSH 切断後もセッションは残ります。

```bash
exit
```

再接続後:

```bash
ssh <user>@<server>
agent-deck list
agent-deck conductor status ops
agent-deck session output conductor-ops
agent-deck session output api-fix
agent-deck session output ui-task
tmux ls
```

## 11. 最初に試すならこの 3 本構成

最初は **Conductor 1 本 + 子セッション 2 本** が分かりやすいです。

### Claude で試す

```bash
agent-deck conductor setup ops --description "Claude conductor"
agent-deck session start conductor-ops

agent-deck launch ~/work/myrepo -t backend  -c claude -m "バックエンドを調査し、不具合修正を開始してください。"
agent-deck launch ~/work/myrepo -t frontend -c claude -m "フロントエンドを調査し、影響を受ける UI を特定してください。"

agent-deck session send conductor-ops "backend と frontend を監視し、状況を要約し、どちらかが waiting または blocked になったら知らせてください。"
```

### Copilot で試す

```bash
agent-deck conductor setup ops --agent copilot --model claude-sonnet-4.6 --allow-all
agent-deck session start conductor-ops

agent-deck launch ~/work/myrepo -t backend  -c "copilot --allow-all --model claude-sonnet-4.6" -m "バックエンドを調査し、不具合修正を開始してください。"
agent-deck launch ~/work/myrepo -t frontend -c "copilot --allow-all --model claude-sonnet-4.6" -m "フロントエンドを調査し、影響を受ける UI を特定してください。"

agent-deck session send conductor-ops "backend と frontend を監視し、状況を要約し、どちらかが waiting または blocked になったら知らせてください。"
```

### Codex で試す

```bash
agent-deck conductor setup ops --agent codex --description "Codex conductor"
agent-deck session start conductor-ops

agent-deck launch ~/work/myrepo -t backend  -c codex -m "バックエンドを調査し、不具合修正を開始してください。"
agent-deck launch ~/work/myrepo -t frontend -c codex -m "フロントエンドを調査し、影響を受ける UI を特定してください。"

agent-deck session send conductor-ops "backend と frontend を監視し、状況を要約し、どちらかが waiting または blocked になったら知らせてください。"
```

役割:

- **conductor-ops**: 監督
- **backend**: バックエンド担当
- **frontend**: フロントエンド担当

## 12. 困ったときに最初に見るコマンド

```bash
agent-deck list
agent-deck conductor status
agent-deck session output conductor-ops
agent-deck session output backend
agent-deck session output frontend
```

この順で見ると、**全体状態 → 監督状態 → 各子セッションの内容** の順に追えます。

## 13. まず本当に試すための最小セット

最小構成なら、**Conductor 1 本 + 子 1 本** で十分です。

### Claude

```bash
agent-deck conductor setup ops
agent-deck session start conductor-ops

agent-deck launch ~/work/myrepo \
  -t test-child \
  -c claude \
  -m "このリポジトリを調査し、何があるかを要約してください。"

agent-deck session send conductor-ops "test-child を監視し、現在の状態を報告してください。"
```

### Copilot

```bash
agent-deck conductor setup ops --agent copilot --model claude-sonnet-4.6 --allow-all
agent-deck session start conductor-ops

agent-deck launch ~/work/myrepo \
  -t test-child \
  -c "copilot --allow-all --model claude-sonnet-4.6" \
  -m "このリポジトリを調査し、何があるかを要約してください。"

agent-deck session send conductor-ops "test-child を監視し、現在の状態を報告してください。"
```

### Codex

```bash
agent-deck conductor setup ops --agent codex
agent-deck session start conductor-ops

agent-deck launch ~/work/myrepo \
  -t test-child \
  -c codex \
  -m "このリポジトリを調査し、何があるかを要約してください。"

agent-deck session send conductor-ops "test-child を監視し、現在の状態を報告してください。"
```

その後に SSH を切り、再接続後に以下を確認します。

```bash
agent-deck list
agent-deck session output conductor-ops
agent-deck session output test-child
```

## 14. ランタイム別の実践メモ

### Claude

- 一番素直に使いやすい標準パターンです
- まず迷ったら `claude` で始めると理解しやすいです

### Copilot

- `-g <group>` より **絶対パスを明示**する方が安全です
- 子セッションは `-c copilot` より **`-c "copilot --allow-all --model ..."`** が安定しやすいです
- `session send` が届かないときは **`tmux send-keys ... $'\r'`** を使います
- 初回ディレクトリアクセス確認を避けたいときは **`--allow-all`** が有効です

必要なら `~/.agent-deck/config.toml` に Copilot の既定値を入れられます。

```toml
[copilot]
default_model = "claude-sonnet-4.6"
allow_all = true
conductor_model = "claude-sonnet-4.6"
conductor_allow_all = true
```

### Codex

- conductor では `--agent codex` を使います
- Codex conductor は `AGENTS.md` ベースで動きます
- 子セッションは `-c codex` を基本に、必要なら追加オプション付きのコマンド文字列にできます

## 15. 関連ドキュメント

- `README.md`: 全体機能の紹介
- `documentation/CONDUCTOR.md`: Conductor の詳細
- `skills/agent-deck/references/cli-reference.md`: CLI の一覧
- `skills/agent-deck/references/config-reference.md`: 設定ファイルの一覧
- `docs/feature-copilot/copilot-session-issues-and-improvements.md`: Copilot 運用時の注意点
