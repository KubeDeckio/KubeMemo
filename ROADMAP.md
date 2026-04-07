# Roadmap

This file tracks the next planned work for KubeMemo after `v0.0.1`.

It is intentionally short and product-focused. The goal is to keep the next release clear rather than turn the roadmap into a dumping ground.

## v0.0.2

The focus for `v0.0.2` should be making KubeMemo feel more production-ready in day-to-day Kubernetes operations.

### Status after v0.0.1 hardening

Some of the original `v0.0.2` ideas were already pulled forward into `v0.0.1`.

- GitOps export/import structure is already much better than the original first draft.
- RBAC-aware errors and capability reporting are already much stronger.
- The TUI already has filtering, detail scrolling, and better navigation than first planned.
- Cluster-backed smoke tests and Kind-gated release validation are already in place.
- Expired runtime memos are already hidden from normal views by default.

That means `v0.0.2` is now more about deepening those areas than introducing them for the first time.

### Goals

- make activity capture more reliable and easier to operate
- improve GitOps workflows beyond basic export/import
- tighten RBAC-aware behavior and user feedback
- polish the TUI and memo rendering for larger real-world memo sets
- improve installation and release confidence with more validation and tests

### Planned work

#### 1. Activity capture hardening

- improve watcher reliability for longer-running sessions
- tighten deduplication behavior and time-window merging
- improve actor/source inference where possible
- extend watched resource coverage and Gateway API handling
- document recommended deployment patterns for always-on capture

#### 2. GitOps workflow depth

- improve durable export layout so it matches a clearer namespace/app/resource structure
- make durable edit/remove flows generate cleaner Git-managed output
- improve GitOps mode detection and user-facing guidance
- make runtime-store safety checks more explicit

Current state:

- structured export/import is already in place
- the main remaining work is better edit/remove Git-managed output

#### 3. RBAC and permissions UX

- improve permission-specific error messages
- make read/write/patch capability reporting more explicit
- improve namespace-scoped and partial-visibility behavior
- surface capability limits more clearly in the TUI and status commands

Current state:

- permission errors and capability summaries are already much improved
- the main remaining work is more TUI visibility/scope messaging and broader real-cluster RBAC validation

#### 4. TUI and rendering polish

- improve large-memo-set navigation and filtering
- add grouping or paging options for busier clusters
- continue improving the memo-board and detail-pane UX
- refine terminal rendering for long content and narrow terminals

Current state:

- filtering, jump keys, help, and detail scrolling are already in place
- the main remaining work is grouping/paging for busier clusters and further density polish

#### 5. Testing and release confidence

- add more cluster-backed integration tests
- add workflow validation for release packaging paths
- add more coverage for annotations, GitOps, and activity capture
- improve release and install smoke-test guidance

Current state:

- cluster-backed smoke tests and Kind-gated release validation are already in place
- the main remaining work is broader activity-capture, GitOps, and packaging verification

## Later

These are good candidates after `v0.0.2`, but they should not block the next release.

- Krew packaging
- richer GitOps repository structure helpers
- optional external integrations such as Confluence or GitHub Wiki export
- deeper KubeBuddy integration hooks
- broader TUI workflows for edit/remove/create inside the interface

## Issue-ready backlog

These are written so they can be turned into GitHub issues with minimal editing.

### v0.0.2 Issue 1

**Title:** Harden activity capture deduplication and long-running watcher behavior

**Summary:** Improve the reliability of foreground and in-cluster activity capture for longer-running sessions and reduce duplicate activity memos during repeated identical changes.

**Acceptance criteria:**

- repeated identical changes inside the configured window do not create duplicate runtime activity memos
- watcher behavior remains stable over longer-running sessions
- docs explain dedupe behavior clearly

### v0.0.2 Issue 2

**Title:** Improve GitOps durable export layout and edit/remove workflows

**Summary:** Make durable memo export/import and file-generation flows better match a practical GitOps repository structure.

**Acceptance criteria:**

- exported durable memos can be organized by namespace, app, or resource path more cleanly
- edit/remove flows generate Git-friendly output
- docs explain the recommended layout and workflow

### v0.0.2 Issue 3

**Title:** Improve RBAC-aware errors and capability reporting

**Summary:** Make KubeMemo clearer when reads, writes, or annotation patches are blocked by permissions.

**Acceptance criteria:**

- commands fail with permission-specific messages instead of generic kubectl failures
- installation/status commands surface readable capability summaries
- TUI indicates partial visibility or scope limitations more clearly

### v0.0.2 Issue 4

**Title:** Add larger-cluster TUI navigation improvements

**Summary:** Improve the TUI for clusters with many memos by adding better filtering, grouping, or paging behavior.

**Acceptance criteria:**

- busy memo lists remain readable and navigable
- filtering works cleanly across namespace, kind, and text
- detail pane remains usable with long memo content

### v0.0.2 Issue 5

**Title:** Expand integration and smoke tests for release confidence

**Summary:** Add stronger test coverage for install, annotations, GitOps, activity capture, and release packaging so future releases are safer to tag.

**Acceptance criteria:**

- cluster-backed integration tests cover the critical install/create/show/remove path
- activity capture has explicit test coverage
- release packaging validation covers the main published artifacts
