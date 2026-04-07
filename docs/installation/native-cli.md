# Native CLI

KubeMemo is distributed as a native Go CLI for macOS, Linux, and Windows.

The native CLI is the primary product surface. It gives you:

- the full `kubememo` command set
- memo-style card rendering
- the interactive TUI
- activity watcher commands
- the same core experience across platforms

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

## Bootstrap the cluster

After the CLI is installed, bootstrap the cluster prerequisites:

```bash
kubememo install --enable-runtime-store --install-rbac --output json
```

Enable the optional in-cluster activity watcher during install:

```bash
kubememo install \
  --enable-runtime-store \
  --install-rbac \
  --enable-activity-capture \
  --activity-capture-image ghcr.io/kubedeckio/kubememo:0.0.1 \
  --output json
```

## Optional next step

If you want the always-on cluster deployment path instead of CLI-driven installation, see the Helm page:

- [Helm chart deployment](helm.md)
