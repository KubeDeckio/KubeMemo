# kubectl Describe View

KubeMemo annotations are designed to make `kubectl describe` more useful without dumping the full memo body into resource metadata.

## Example

After durable memos are attached to a Deployment, a `kubectl describe` can show lightweight memo metadata like:

```text
Annotations:
  notes.kubememo.io/enabled:      true
  notes.kubememo.io/has-note:     true
  notes.kubememo.io/note-count:   3
  notes.kubememo.io/runtime-count: 1
  notes.kubememo.io/view:         kubememo show --kind Deployment --namespace prod --name orders-api
```

## Why this works

- operators immediately know memo context exists
- the annotation stays small and readable
- the `view` command points to the real source of truth
- the resource is not bloated with full markdown note content

## Design principle

Annotations are for discoverability.

The actual durable and runtime memo content still lives in the KubeMemo CRDs, where it can be rendered properly in the CLI and TUI.
