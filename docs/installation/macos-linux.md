# macOS and Linux Installation

KubeMemo is distributed as a native Go CLI for macOS and Linux.

## Homebrew

Once the Homebrew tap is updated, install with:

```bash
brew install KubeDeckio/tap/kubememo
```

Upgrade with:

```bash
brew upgrade kubememo
```

## GitHub Releases

You can also install directly from release assets on GitHub:

1. Download the archive for your platform from the latest release.
2. Extract `kubememo`.
3. Move it somewhere on your `PATH`.

Example for macOS or Linux after extracting:

```bash
chmod +x kubememo
sudo mv kubememo /usr/local/bin/
```

## Verify the install

```bash
kubememo version --output json
kubememo status --output json
```

## Optional in-cluster activity capture with Helm

The Helm chart deploys the optional activity-capture component with secure defaults:

- non-root container
- dropped Linux capabilities
- `RuntimeDefault` seccomp profile
- read-only root filesystem
- explicit CPU and memory requests and limits

Install the chart into `kubememo-runtime`:

```bash
helm upgrade --install kubememo ./charts/kubememo \
  --namespace kubememo-runtime \
  --create-namespace \
  --set activityCapture.enabled=true \
  --set image.repository=ghcr.io/kubedeckio/kubememo \
  --set image.tag=0.0.1
```

This installs the CRDs and the optional in-cluster watcher. By default, the watcher writes runtime activity memos into the release namespace.

## Next step

After the CLI is installed, bootstrap the cluster prerequisites:

```bash
kubememo install --enable-runtime-store --install-rbac --output json
```
