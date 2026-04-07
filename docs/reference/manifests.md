# Example Manifests

KubeMemo can be used through the CLI and PowerShell wrapper, but the CRDs are also simple enough to work with directly through `kubectl` or GitOps-managed YAML.

Use these examples when you want to:

- test the CRDs quickly with `kubectl apply -f`
- store durable memos in Git
- understand the shape of `Memo` and `RuntimeMemo` resources

## Durable memo example

This is the long-lived curated memo shape you would normally keep in Git for ownership, runbooks, warnings, or maintenance guidance.

```yaml
apiVersion: notes.kubememo.io/v1alpha1
kind: Memo
metadata:
  name: orders-api-warmup
  namespace: prod
  labels:
    notes.kubememo.io/type: warning
    app.kubernetes.io/name: orders-api
spec:
  title: Orders API warm-up behavior
  summary: Expected transient 502s after deployment
  content: |
    Ignore failures for up to 3 minutes after rollout.
    Investigate only if the error rate continues climbing after the warm-up window.
  format: markdown
  noteType: warning
  severity: info
  owner:
    team: platform-apps
    contact: "@platform-apps"
  target:
    mode: resource
    apiVersion: apps/v1
    kind: Deployment
    namespace: prod
    name: orders-api
    appRef:
      name: orders-api
      instance: prod
  tags:
    - deploy
    - warmup
  source:
    type: git
    git:
      repo: KubeDeckio/ops
      path: ops/kubememo/apps/orders-api/warmup.yaml
      revision: main
status:
  state: active
  expired: false
```

Apply it directly:

```bash
kubectl apply -f memo-orders-api.yaml
```

## Runtime memo example

This is the short-lived runtime shape used for incident notes, temporary warnings, or auto-captured activity breadcrumbs.

```yaml
apiVersion: runtime.notes.kubememo.io/v1alpha1
kind: RuntimeMemo
metadata:
  name: orders-api-scale-change-20260407-120000
  namespace: kubememo-runtime
  labels:
    notes.kubememo.io/type: activity
    app.kubernetes.io/name: orders-api
spec:
  title: Manual scale change detected
  summary: Replicas changed from 2 to 5
  content: |
    Manual scale change recorded during incident review.
  format: markdown
  noteType: activity
  temporary: true
  severity: info
  target:
    mode: resource
    apiVersion: apps/v1
    kind: Deployment
    namespace: prod
    name: orders-api
    appRef:
      name: orders-api
      instance: prod
  source:
    type: auto
    generator: activity-capture
    confidence: medium
  activity:
    action: scale
    fieldPath: spec.replicas
    oldValue: "2"
    newValue: "5"
    actor: kubernetes-admin
    actorType: user
    detectedAt: "2026-04-07T12:00:00Z"
  createdAt: "2026-04-07T12:00:00Z"
  expiresAt: "2026-04-08T12:00:00Z"
status:
  state: active
  expired: false
```

Apply it directly:

```bash
kubectl apply -f runtime-memo-orders-api.yaml
```

## Namespace-targeted durable memo

Use namespace targeting when the guidance applies to the whole namespace instead of one resource.

```yaml
apiVersion: notes.kubememo.io/v1alpha1
kind: Memo
metadata:
  name: payments-prod-maintenance-window
  namespace: payments-prod
spec:
  title: Namespace maintenance window
  summary: Avoid manual rollouts during the nightly reconciliation window
  content: |
    GitOps reconciliation and database maintenance happen between 01:00 and 01:30 UTC.
  format: markdown
  noteType: maintenance
  severity: info
  target:
    mode: namespace
    namespace: payments-prod
status:
  state: active
  expired: false
```

## GitOps usage

For GitOps clusters:

- keep durable `Memo` objects in Git
- keep runtime `RuntimeMemo` objects outside Git reconciliation scope
- avoid putting temporary incident context into durable manifests

Recommended durable layout:

```text
ops/
  kubememo/
    namespaces/
      payments-prod.yaml
    apps/
      orders-api/
        warning-warmup.yaml
    resources/
      prod/
        deployment-orders-api.yaml
```

## Quick verification

```bash
kubectl get memos -A
kubectl get runtimememos -A
kubectl describe deployment orders-api -n prod
kubememo show --kind Deployment --namespace prod --name orders-api
```
