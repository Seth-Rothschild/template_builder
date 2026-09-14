#!/bin/sh
set -e

os=$(uname -s | tr '[:upper:]' '[:lower:]')
machine=$(uname -m)

if [ "$machine" = "x86_64" ]; then
    arch="amd64"
elif [ "$machine" = "arm64" ] || [ "$machine" = "aarch64" ]; then
    arch="arm64"
else
    arch="$machine"
fi

if command -v go >/dev/null 2>&1; then
    echo "go already installed: $(go version)"
else
    echo "installing latest Go..."
    go_version=$(curl -fsSL "https://go.dev/VERSION?m=text" | head -n1)
    curl -fsSL "https://go.dev/dl/$go_version.$os-$arch.tar.gz" -o /tmp/go.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    echo "Go installed to /usr/local/go/bin."
fi

if command -v tectonic >/dev/null 2>&1; then
    echo "tectonic already installed: $(tectonic --version)"
else
    echo "installing latest tectonic..."
    curl --proto '=https' --tlsv1.2 -fsSL https://drop-sh.fullyjustified.net | sh
    echo "moving tectonic to /usr/local/bin, this may prompt for your password..."
    sudo mv ./tectonic /usr/local/bin/tectonic
    echo "tectonic installed to /usr/local/bin/tectonic."
fi

if command -v pandoc >/dev/null 2>&1; then
    echo "pandoc already installed: $(pandoc --version | head -n1)"
else
    echo "installing latest pandoc..."
    latest_release_url=$(curl -fsSL -o /dev/null -w '%{url_effective}' https://github.com/jgm/pandoc/releases/latest)
    pandoc_version=$(basename "$latest_release_url")

    if [ "$os" = "darwin" ]; then
        if [ "$arch" = "amd64" ]; then
            mac_arch="x86_64"
        else
            mac_arch="$arch"
        fi
        curl -fsSL "https://github.com/jgm/pandoc/releases/download/$pandoc_version/pandoc-$pandoc_version-$mac_arch-macOS.pkg" -o /tmp/pandoc.pkg
        sudo installer -pkg /tmp/pandoc.pkg -target /
        rm /tmp/pandoc.pkg
    else
        curl -fsSL "https://github.com/jgm/pandoc/releases/download/$pandoc_version/pandoc-$pandoc_version-linux-$arch.tar.gz" -o /tmp/pandoc.tar.gz
        tar -xzf /tmp/pandoc.tar.gz -C /tmp
        echo "moving pandoc to /usr/local/bin, this may prompt for your password..."
        sudo mv "/tmp/pandoc-$pandoc_version/bin/pandoc" /usr/local/bin/pandoc
        rm -rf /tmp/pandoc.tar.gz "/tmp/pandoc-$pandoc_version"
    fi
    echo "pandoc installed to /usr/local/bin/pandoc."
fi

case ":$PATH:" in *":/usr/local/go/bin:"*) ;; *) echo "warning: /usr/local/go/bin is not in your PATH" ;; esac
