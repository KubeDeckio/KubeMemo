# Annotations

KubeMemo uses annotations to make memo context discoverable from normal Kubernetes workflows such as `kubectl describe`.

Annotations are not the source of truth. They are only a lightweight hint that memo context exists and how to open it.

## Why KubeMemo uses annotations

Operators often inspect the resource before they open a dedicated tool.

That means `kubectl describe` is a useful place to answer:

- does this resource have memo context?
- how many memos exist?
- is there runtime context as well?
- what command should I run to read the full story?

## Design rules

KubeMemo keeps annotations intentionally small:

- never store full markdown memo content
- do not try to list every memo on a busy resource
- prefer counts over long lists
- only include a short summary when it is unambiguous
- always point back to the real source of truth

## Annotation keys

These are the main keys KubeMemo uses:

- `notes.kubememo.io/enabled`
- `notes.kubememo.io/has-note`
- `notes.kubememo.io/note-count`
- `notes.kubememo.io/runtime-count`
- `notes.kubememo.io/runtime-enabled`
- `notes.kubememo.io/summary`
- `notes.kubememo.io/view`

## What each key means

### `notes.kubememo.io/enabled`

Marks the target as note-enabled for discovery and activity-capture logic.

### `notes.kubememo.io/has-note`

Signals that memo context exists on this target.

### `notes.kubememo.io/note-count`

Shows how many durable memos are attached to the target.

### `notes.kubememo.io/runtime-count`

Shows how many runtime memos currently exist for the target.

### `notes.kubememo.io/runtime-enabled`

Signals that runtime memo behavior is enabled or relevant for this target.

### `notes.kubememo.io/summary`

Stores one short summary string only when KubeMemo can do that cleanly.

KubeMemo does not try to concatenate many memo summaries into one annotation.

### `notes.kubememo.io/view`

Stores the command to open the full memo view for the target, for example:

```text
kubememo show --kind Deployment --namespace prod --name orders-api
```

## How summary selection works

The `summary` annotation is intentionally conservative.

KubeMemo uses these rules:

- prefer durable memo summaries over runtime memo summaries
- when you just created or updated a memo, that memo gets priority
- only write a summary when the result is still clear and useful
- truncate long summaries so the annotation stays readable

That means the annotation stays helpful without turning into metadata spam.

## Example

After attaching durable and runtime memos to a Deployment, `kubectl describe` might show:

```text
Annotations:
  notes.kubememo.io/enabled:         true
  notes.kubememo.io/has-note:        true
  notes.kubememo.io/note-count:      2
  notes.kubememo.io/runtime-count:   1
  notes.kubememo.io/summary:         Expected transient 502s after deployment
  notes.kubememo.io/view:            kubememo show --kind Deployment --namespace prod --name orders-api
```

This is enough to tell an operator:

- memo context exists
- there are both durable and runtime notes
- one short summary is available
- the full view command is right there

## Automatic sync behavior

In non-GitOps direct-write mode, KubeMemo can sync annotations automatically for:

- exact resource targets
- namespace targets

That means creating, updating, or removing memos can keep the target annotations current without a separate manual step.

## GitOps behavior

Annotations need to respect the same operational rules as the rest of KubeMemo.

In GitOps environments:

- durable memo workflows may be Git-managed instead of directly patched in-cluster
- runtime memo annotations should still stay lightweight
- automatic annotation patching should not fight reconciliation policy

If a cluster or workflow does not allow direct resource patching, annotations should be treated as optional discoverability metadata, not required state.

## Why KubeMemo does not store full memo bodies in annotations

That would create the exact problems annotations should avoid:

- large noisy resource metadata
- duplicated source of truth
- poor readability in `kubectl describe`
- harder GitOps diffs
- more churn from temporary runtime notes

The memo body belongs in the KubeMemo CRDs, where it can be rendered properly in the CLI and TUI.

## Related pages

- [kubectl Describe View](describe-view.md)
- [Durable vs Runtime Memos](stores.md)
- [Activity Capture](activity-capture.md)
- [GitOps Behavior](gitops.md)
