# Helm Chart

KubeMemo ships an optional Helm chart for teams that want the always-on in-cluster activity watcher.

This is the right path when you want activity breadcrumbs to continue working without a user keeping a terminal open.

## What the chart deploys

The chart can install:

- the KubeMemo CRDs
- the runtime namespace
- the activity-capture Deployment
- the required ServiceAccount and RBAC
- optional PodDisruptionBudget and NetworkPolicy

## Security defaults

The activity-capture workload is hardened by default:

- non-root container
- explicit `runAsUser` and `runAsGroup`
- `allowPrivilegeEscalation: false`
- dropped Linux capabilities
- `RuntimeDefault` seccomp profile
- read-only root filesystem
- resource requests and limits

## Install the chart

```bash
helm upgrade --install kubememo ./charts/kubememo \
  --namespace kubememo-runtime \
  --create-namespace \
  --set activityCapture.enabled=true \
  --set image.repository=ghcr.io/kubedeckio/kubememo \
  --set image.tag=0.0.1
```

## What the in-cluster watcher does

When enabled, the activity-capture Deployment:

- watches selected Kubernetes resources
- detects meaningful changes instead of generic status churn
- writes runtime `activity` memos into the runtime store
- deduplicates repeated identical changes inside the configured window

## Values you are likely to care about

```yaml
image:
  repository: ghcr.io/kubedeckio/kubememo
  tag: 0.0.1

activityCapture:
  enabled: true

runtimeNamespace:
  create: true
  name: kubememo-runtime
```

## Helm or CLI install flags?

Use Helm when:

- you want an always-on cluster deployment
- your team already manages operational add-ons with Helm
- you want the watcher to survive local terminal sessions

Use `kubememo install --enable-activity-capture` when:

- you want KubeMemo to bootstrap the same capability directly
- you are using the CLI as the primary installation experience

## Next steps

- [Activity capture model](../concepts/activity-capture.md)
- [Annotations](../concepts/annotations.md)
- [GitOps behavior](../concepts/gitops.md)
