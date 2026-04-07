# CLI Overview

KubeMemo is now a Go-native CLI. The PowerShell module wraps this binary rather than re-implementing the product.

## Examples

```bash
kubememo get --namespace prod --output json
kubememo show --kind Deployment --namespace prod --name orders-api
kubememo tui
```

## Human-facing defaults

- `kubememo show` includes runtime memos by default
- `kubememo find` includes runtime memos by default
- `kubememo tui` includes runtime memos by default

`kubememo get` remains durable-first because it is the more automation-oriented command.

## Output modes

- JSON for scripting
- Memo-style card rendering for terminal reading
- TUI for interactive browsing
