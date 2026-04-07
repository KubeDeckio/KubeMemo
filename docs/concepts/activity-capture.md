# Activity Capture

Activity capture is one of KubeMemo’s core features.

Its job is simple: watch important Kubernetes changes and turn them into readable runtime memos that explain what changed.

## Why it exists

Operators often see the result of a change without seeing the story behind it.

Examples:

- a Deployment was manually scaled
- an image was changed during incident response
- a Service type was changed for a quick workaround
- a route or gateway object was updated under pressure

KubeMemo records those as runtime `activity` memos so the context can be read later in the CLI, TUI, or `kubectl describe` flow.

## Two ways to run it

### Foreground watcher

Use the CLI or PowerShell wrapper:

```bash
kubememo watch-activity --namespace prod --kind Deployment
```

```powershell
Start-KubeMemoActivityCapture -Namespace prod -Kind Deployment
```

This mode runs until you stop it.

### In-cluster watcher

Use the optional Helm chart or install flag:

```bash
kubememo install --enable-activity-capture --enable-runtime-store --install-rbac
```

or:

- [Helm chart deployment](../installation/helm.md)

This mode is the better choice when you want activity capture to be always on.

## What it watches

The first cut focuses on meaningful operator-facing resources:

- Deployment
- StatefulSet
- DaemonSet
- Service
- Ingress
- HorizontalPodAutoscaler
- Gateway
- HTTPRoute

## What it captures

The watcher looks for high-value change types, including:

- scale changes
- image changes
- resource request and limit changes
- Service type changes
- Ingress changes
- Gateway listener changes
- HTTPRoute backend and match changes
- node selector changes
- toleration changes

## What it avoids

The watcher is intentionally selective.

It does not try to record:

- pod churn
- status-only updates
- generic metadata noise
- every controller-generated change

## Where it writes

Activity capture always writes to the runtime store, never the durable store.

That means:

- activity breadcrumbs stay temporary and operational
- durable memos remain curated
- GitOps clusters can keep durable notes in Git while runtime activity stays live in-cluster

## Deduplication

KubeMemo deduplicates repeated identical changes inside a short time window.

The goal is to avoid writing the same activity memo again and again if the same exact change is observed repeatedly.

## Notes-enabled behavior

Activity capture is designed to focus on targets that are already relevant to KubeMemo.

That includes resources with:

- existing memos
- memo enablement annotations
- inherited note-enabled context from policy or namespace

## Result

The output is a runtime memo with:

- note type `activity`
- the target resource
- the changed field
- old and new values
- source metadata showing the note was auto-generated

That runtime memo then appears in:

- `kubememo show`
- `kubememo tui`
- `Get-KubeMemo -IncludeRuntime`
