# Install cc-bash-guard

Use Homebrew or mise for most installs. Manual GitHub Releases and Go builds
are available for environments that need explicit artifact handling or a local
Go toolchain build.

## Homebrew

```sh
brew tap tasuku43/cc-bash-guard
brew install cc-bash-guard
cc-bash-guard init
cc-bash-guard verify
```

The formula in
[`tasuku43/homebrew-cc-bash-guard`](https://github.com/tasuku43/homebrew-cc-bash-guard)
pins SHA-256 checksums against GitHub Releases archives.

## mise

```sh
mise use -g github:tasuku43/cc-bash-guard@latest
cc-bash-guard init
cc-bash-guard verify
```

## GitHub Releases

GitHub Releases publish prebuilt `tar.gz` archives for macOS and Linux on
arm64 and amd64.

```sh
TAG=<tag>       # replace with the release tag you want
OS=macos        # macos or linux
ARCH=arm64      # arm64 or x64
ARCHIVE="cc-bash-guard_${TAG}_${OS}_${ARCH}.tar.gz"

curl -LO "https://github.com/tasuku43/cc-bash-guard/releases/download/${TAG}/${ARCHIVE}"
curl -LO "https://github.com/tasuku43/cc-bash-guard/releases/download/${TAG}/checksums.txt"
```

Verify the SHA-256 checksum before installing:

```sh
# macOS
grep "  ${ARCHIVE}$" checksums.txt | shasum -a 256 -c -

# Linux
grep "  ${ARCHIVE}$" checksums.txt | sha256sum -c -
```

Put the binary on `PATH`:

```sh
mkdir -p "$HOME/.local/bin"
tar -xzf "$ARCHIVE" cc-bash-guard
install -m 0755 cc-bash-guard "$HOME/.local/bin/cc-bash-guard"
cc-bash-guard init
cc-bash-guard verify
```

## Build From Source

For Go toolchain users:

```sh
go install github.com/tasuku43/cc-bash-guard/cmd/cc-bash-guard@latest
cc-bash-guard init
cc-bash-guard verify
```

The module requires Go 1.25 or newer. Make sure your Go install directory is on
`PATH`; by default that is usually `$(go env GOPATH)/bin`.

## Upgrade

Homebrew:

```sh
brew update
brew upgrade cc-bash-guard
cc-bash-guard verify
```

mise:

```sh
mise use -g github:tasuku43/cc-bash-guard@latest
cc-bash-guard verify
```

Go-installed binary:

```sh
go install github.com/tasuku43/cc-bash-guard/cmd/cc-bash-guard@latest
cc-bash-guard verify
```

For GitHub Releases, download the newer archive, verify it against that
release's `checksums.txt`, replace the binary on `PATH`, then run:

```sh
cc-bash-guard version
cc-bash-guard verify
```

Verified artifacts include an evaluation semantics version and the resolved
policy inputs. Regenerate them with `cc-bash-guard verify` after upgrading.
