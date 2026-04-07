# Annotations

KubeMemo can annotate target resources so `kubectl describe` shows that memo context exists.

## Annotation model

KubeMemo uses lightweight discovery annotations rather than storing full note bodies on the resource.

Typical keys:

- `notes.kubememo.io/enabled`
- `notes.kubememo.io/has-note`
- `notes.kubememo.io/note-count`
- `notes.kubememo.io/runtime-count`
- `notes.kubememo.io/summary`
- `notes.kubememo.io/view`

## Design rules

- Do not store full markdown memo content in annotations.
- Prefer counts over listing every memo reference once a resource has many notes.
- Only include a short summary when it is unambiguous.
- Include a `kubememo show ...` command hint so the real source of truth is easy to open.

## Automatic behavior

For direct writes in non-GitOps mode, KubeMemo can sync these annotations automatically for resource and namespace targets.
