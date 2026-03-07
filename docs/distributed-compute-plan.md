# Distributed Compute for Agent-Deck

## Architecture

Conductors run locally on the main host. Compute is distributed by delegating session creation and management to remote SSH-accessible machines, each running their own `agent-deck` binary and tmux.

```
┌─────────────────────────────────────┐
│         Local Host (this Mac)       │
│                                     │
│  ┌──────────┐    ┌───────────────┐  │
│  │ TUI      │    │ Conductor     │  │
│  │ (home.go)│    │ (Claude in    │  │
│  │          │    │  local tmux)  │  │
│  └────┬─────┘    └───────┬───────┘  │
│       │                  │          │
│       │ attach via SSH   │ CLI via  │
│       │                  │ SSH      │
└───────┼──────────────────┼──────────┘
        │                  │
   ┌────┴──────────────────┴──────┐
   │      Remote Host (SSH)       │
   │                              │
   │  agent-deck (remote)         │
   │  ├── tmux sessions           │
   │  ├── claude/gemini agents    │
   │  └── repos & worktrees       │
   └──────────────────────────────┘
```

## Current State

Partial SSH support already exists:
- `SSHRunner` (session/ssh.go) can list, attach, and run commands on remote agent-deck instances
- `Instance` has `SSHHost` and `SSHRemotePath` fields
- `wrapForSSH` wraps commands in SSH invocations with ControlMaster
- `RemoteConfig` in config.toml defines remote hosts (`[remotes.mybox]`)
- Remote sessions appear read-only in the TUI sidebar (fetched every 30s)
- Conductors run locally and use the `agent-deck` CLI to manage sessions

## Key Design Decisions

1. **Remote must run its own agent-deck** — the local side delegates via CLI over SSH. Simpler and more robust than managing remote tmux directly.
2. **Conductor stays local** — conductors are meta-agents that orchestrate. They run locally in local tmux and issue CLI commands that may target remote hosts.
3. **`--remote` flag is the primary interface** — maps to a named remote in config.toml. Per-session `SSHHost` is the storage mechanism for tracking which sessions are remote.
4. **ControlMaster for SSH performance** — already implemented. Persistent SSH connections avoid handshake overhead.

## Phases

### Phase 1: Conductor remote dispatch (start here)

Highest leverage, least work. ~100 lines of Go + template update. No TUI changes.

**Goal**: Conductors can create and manage sessions on remote hosts.

**Changes**:
- Add `--remote <name>` global CLI flag to `main.go`
- When set, build an `SSHRunner` from the named remote in config.toml and delegate the entire command via `SSHRunner.Run()`
- Update conductor CLAUDE.md template to document remote CLI patterns:
  ```
  agent-deck -p <PROFILE> --remote <name> launch <path> ...
  agent-deck -p <PROFILE> --remote <name> session send <id> "msg"
  agent-deck -p <PROFILE> --remote <name> session output <id> -q
  ```
- Conductor template should list available remotes and their purpose

### Phase 2: Remote session creation from TUI

**Goal**: The `[n]` new session dialog can target a remote host.

**Changes**:
- New dialog field: "Host" dropdown (local / configured remotes from config.toml)
- When a remote host is selected:
  - Path field browses/autocompletes remote filesystem (via `ssh ls` or `ssh find`)
  - Tool selection stays the same
  - Worktree options work if remote repo supports them
- On submit: `SSHRunner.Run(ctx, "launch", path, "-t", title, "-c", tool, "-g", group)` instead of local instance creation
- Trigger remote session re-fetch so it appears immediately
- Store `SSHHost` on the Instance

### Phase 3: Unified session view

**Goal**: Remote sessions appear inline with local sessions, not in a separate sidebar.

**Changes**:
- Remote sessions get first-class `Instance` objects with `SSHHost` set
- Status polling via periodic `SSHRunner.FetchSessions()` updates status
- Group tree supports mixed local+remote sessions (SSH badge on remote items)
- Preview capture for remote sessions: `ssh host 'agent-deck session output <id> -q'`
- Notification bar includes remote sessions

### Phase 4: Remote worktree support

**Goal**: Create worktrees on remote repos.

**Changes**:
- Worktree creation/deletion commands dispatched via SSH
- Post-create and post-delete hooks run on the remote host (scripts live on remote)
- Branch listing fetched via SSH for worktree dialog

### Phase 5: Resource management

**Goal**: Monitor and manage remote compute resources.

**Changes**:
- Health check for remote hosts (SSH connectivity, disk space, running containers)
- Remote session count limits per host (configurable)
- Automatic cleanup of dead remote sessions
- Resource usage display in TUI (CPU/memory from remote)

## Key Files

| File | Role |
|---|---|
| `internal/session/ssh.go` | SSHRunner — remote command execution, attach, fetch |
| `internal/session/instance.go` | Instance.SSHHost, wrapForSSH, prepareCommand |
| `internal/session/userconfig.go` | RemoteConfig struct, config.toml parsing |
| `internal/session/conductor.go` | Conductor settings and metadata |
| `internal/session/conductor_templates.go` | CLAUDE.md template (needs remote CLI docs) |
| `internal/ui/home.go` | TUI — remote session display, attach, new dialog |
| `cmd/agent-deck/main.go` | CLI entry point (add --remote flag here) |
