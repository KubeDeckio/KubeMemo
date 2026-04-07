# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.1] - 2026-04-07

### Added

* **Go-first KubeMemo CLI and PowerShell wrapper**
  * Added the native `kubememo` Go CLI as the primary product surface.
  * Added a thin PowerShell wrapper module that shells out to the Go binary while preserving PowerShell-friendly command names and object-returning behavior where appropriate.
  * Added version reporting through the native CLI and PowerShell wrapper.

* **Durable and runtime memo stores**
  * Added the durable memo CRD backed by `memos.notes.kubememo.io`.
  * Added the runtime memo CRD backed by `runtimememos.runtime.notes.kubememo.io`.
  * Added install, update, uninstall, validation, CRUD, search, render, export, import, and cleanup flows for the memo stores.
  * Added normalized durable/runtime memo output so read and search flows can work across both stores.

* **Terminal-first operator experience**
  * Added memo-style CLI card rendering for durable and runtime memos.
  * Added the interactive TUI memo board with filtering, runtime visibility, and memo detail rendering.
  * Added colorized note styling and memo-focused terminal layouts for human-facing commands.

* **GitOps-aware and annotation-aware workflows**
  * Added GitOps-aware durable memo behavior and export/import flows.
  * Added lightweight target annotations for memo discovery in `kubectl describe`.
  * Added count-based annotation summaries and `kubememo show ...` command hints on targets.

* **Activity capture**
  * Added watcher-based runtime activity auto-capture for high-value Kubernetes changes.
  * Added support for scale, image, resource, service, ingress, Gateway, and HTTPRoute change detection.
  * Added optional always-on in-cluster activity capture deployment support.

* **Packaging and deployment**
  * Added multi-platform Go binary release packaging for macOS, Linux, and Windows on `amd64` and `arm64`.
  * Added multi-architecture GHCR container publishing for the optional in-cluster watcher.
  * Added an optional Helm chart for the in-cluster activity-capture deployment.
  * Added PowerShell Gallery publishing workflow support.
  * Added Homebrew tap update workflow support.

### Docs

* Added the MkDocs site for KubeMemo with Kubedeck-aligned styling, homepage override, and branded docs structure.
* Added install documentation for native CLI, PowerShell, Windows, and Helm paths.
* Added docs for activity capture, annotations, GitOps behavior, CRDs, TUI usage, and `kubectl describe` discovery.
* Added changelog syncing into the docs site build so the docs changelog page is generated from the repository root changelog.
