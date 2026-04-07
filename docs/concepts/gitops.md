# GitOps Behavior

KubeMemo is GitOps-aware.

## Standard clusters

- Direct durable writes allowed
- Direct runtime writes allowed
- Annotation sync can happen automatically on direct writes

## GitOps clusters

Durable notes should usually be treated as Git-managed. KubeMemo supports manifest/file generation for those workflows rather than assuming direct cluster mutation is always safe.

Runtime memos are only safe when the runtime store is outside GitOps reconciliation scope.

If you want concrete YAML examples for Git-managed durable memos and cluster-managed runtime memos, see:

- [Example manifests](../reference/manifests.md)

## Export layout

`Export-KubeMemo` now writes durable memos into a more practical structure instead of a flat directory:

```text
ops/
  kubememo/
    namespaces/
      payments-prod/
        maintenance-window.yaml
    apps/
      orders-api/
        rollout-runbook.yaml
    resources/
      prod/
        deployment-orders-api/
          warning-warmup.yaml
    runtime/
      kubememo-runtime/
        orders-api-scale-change.yaml
```

This keeps durable memos easier to review in Git while still allowing runtime memo export when you explicitly include it.

## Practical rule

- Durable store: curated source of truth
- Runtime store: live operational context

Do not push temporary incident notes into durable Git-managed memo objects.
