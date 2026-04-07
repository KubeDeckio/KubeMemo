# CRDs

## Durable memo CRD

- Kind: `Memo`
- Group: `notes.kubememo.io`
- Version: `v1alpha1`
- Resource: `memos`
- Short names: `km`, `memo`

## Runtime memo CRD

- Kind: `RuntimeMemo`
- Group: `runtime.notes.kubememo.io`
- Version: `v1alpha1`
- Resource: `runtimememos`
- Short names: `kmr`, `rmemo`

## Examples

```bash
kubectl get memos -A
kubectl get runtimememos -A
kubectl get km -A
kubectl get kmr -A
```

For full YAML examples for `Memo` and `RuntimeMemo`, see:

- [Example manifests](manifests.md)
