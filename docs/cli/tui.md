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

## Navigation

```text
[Arrows]/[j][k] move
[PgUp]/[PgDn] or [u][d] scroll
[/] text filter
[:] switch view
[f] namespace filter
[c] kind filter
[a] add temporary memo
[r] refresh
[q] quit
```

## Why the TUI exists

The TUI is for discovery and incident-time reading. It complements the CLI rather than replacing it.
