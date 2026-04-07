# Durable vs Runtime Memos

KubeMemo uses a split-store model.

## Durable memos

- Resource: `memos.notes.kubememo.io`
- Purpose: curated operational memory
- Examples: ownership, warnings, runbooks, suppressions, maintenance notes

## Runtime memos

- Resource: `runtimememos.runtime.notes.kubememo.io`
- Purpose: temporary or live operational context
- Examples: incidents, activity breadcrumbs, short-lived advisories, handover notes

## Why the split matters

In GitOps clusters, durable memos often belong in Git. Runtime memos need a separate non-reconciled live store so operators can still write temporary context without Git fighting those updates.
