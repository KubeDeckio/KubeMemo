# TUI

KubeMemo ships with an interactive terminal UI for browsing memos across namespaces, kinds, and runtime state.

## Launch

```bash
kubememo tui
```

PowerShell wrapper:

```powershell
Open-KubeMemoTui
```

## Defaults

- Includes runtime memos by default
- Auto-refreshes by default
- Supports namespace, kind, and text filtering
- Uses a split memo-board layout with detail pane rendering
- Sorts by recent memo activity by default
- Shows current access posture in the header and status line so RBAC limitations are visible while browsing

## Navigation

```text
[Arrows]/[j][k] move
[g]/[G] jump top/end
[PgUp]/[PgDn] or [u][d] scroll
[/] text filter
[n] namespace filter
[c] kind filter
[a]/[m]/[t] all/durable/runtime
[s] cycle sort
[[]/[]] previous/next page
[x] clear filters
[r] refresh
[?] help
[q] quit
```

## Why the TUI exists

The TUI is for discovery and incident-time reading. It complements the CLI rather than replacing it.
