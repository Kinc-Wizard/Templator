#!/usr/bin/env bash

set -euo pipefail

echo "==> Templator dependency installer"
echo "This will install cross-compilers and toolchains to build C, C#, and Rust Windows payloads."

# Detect package manager
detect_pkg_mgr() {
  if command -v apt-get >/dev/null 2>&1; then echo apt; return; fi
  if command -v pacman >/devnull 2>&1; then echo pacman; return; fi
  if command -v dnf >/dev/null 2>&1; then echo dnf; return; fi
  if command -v zypper >/dev/null 2>&1; then echo zypper; return; fi
  if command -v apk >/dev/null 2>&1; then echo apk; return; fi
  echo unknown
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    return 1
  fi
}

PKG_MGR=$(detect_pkg_mgr)

echo "Detected package manager: ${PKG_MGR}"

case "$PKG_MGR" in
  apt)
    sudo apt-get update -y
    sudo apt-get install -y \
      curl ca-certificates build-essential pkg-config \
      mingw-w64 gcc-mingw-w64 binutils-mingw-w64 \
      mono-complete
    ;;
  pacman)
    sudo pacman -Sy --noconfirm \
      curl ca-certificates base-devel pkgconf \
      mingw-w64-gcc \
      mono
    ;;
  dnf)
    sudo dnf install -y \
      curl ca-certificates @"Development Tools" pkgconf-pkg-config \
      mingw64-gcc mingw32-gcc \
      mono-devel
    ;;
  zypper)
    sudo zypper refresh
    sudo zypper install -y \
      curl ca-certificates -t pattern devel_basis pkg-config \
      mingw64-cross-gcc mingw32-cross-gcc \
      mono-complete || true
    ;;
  apk)
    sudo apk update
    sudo apk add \
      curl ca-certificates build-base pkgconf \
      mingw-w64-gcc \
      mono
    ;;
  *)
    echo "Unsupported or undetected package manager. Please install manually:" >&2
    echo "- MinGW-w64 cross compilers (x86 and x64)" >&2
    echo "- Mono (C# compiler and reference assemblies)" >&2
    echo "- curl, build essentials, pkg-config" >&2
    ;;
esac

echo "\n==> Ensuring Rust toolchain is installed"
if ! command -v cargo >/dev/null 2>&1; then
  echo "Installing Rust via rustup (non-interactive)"
  curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
  # shellcheck disable=SC1091
  source "$HOME/.cargo/env"
else
  echo "Rust is already installed: $(rustc --version || true)"
fi

echo "\n==> Adding Rust Windows GNU targets"
rustup target add x86_64-pc-windows-gnu i686-pc-windows-gnu || true

echo "\n==> Verifying required binaries"
MISSING=0
for bin in x86_64-w64-mingw32-gcc i686-w64-mingw32-gcc; do
  if ! command -v "$bin" >/dev/null 2>&1; then
    echo "[!] Not found: $bin"
    MISSING=1
  else
    echo "[OK] $bin -> $($bin -dumpmachine 2>/dev/null || echo present)"
  fi
done

if command -v mono >/dev/null 2>&1; then
  echo "[OK] mono $(mono --version | head -n1)"
elif command -v mcs >/dev/null 2>&1; then
  echo "[OK] mcs $(mcs --version 2>&1 | head -n1)"
else
  echo "[!] Mono C# compiler not found (mono/mcs)"; MISSING=1
fi

if command -v cargo >/dev/null 2>&1; then
  echo "[OK] cargo $(cargo --version)"
else
  echo "[!] cargo not found in PATH"; MISSING=1
fi

echo "\n==> Validating MinGW library paths (for -L)"
LIB_X64="/usr/x86_64-w64-mingw32/lib"
LIB_X86="/usr/i686-w64-mingw32/lib"
test -d "$LIB_X64" && echo "[OK] $LIB_X64" || echo "[!] Missing $LIB_X64"
test -d "$LIB_X86" && echo "[OK] $LIB_X86" || echo "[!] Missing $LIB_X86"

if [ "$MISSING" -ne 0 ]; then
  echo "\nSome components are missing. Please review the warnings above and install them for your distro."
  exit 1
fi

chmod +x tools/native/Supernova/Supernova
chmod +x tools/native/Astral-PE

echo "\nAll set! You should be able to compile C, C#, and Rust templates."
echo "- C uses: i686-w64-mingw32-gcc / x86_64-w64-mingw32-gcc"
echo "- C# uses: mono/mcs with .NET 4.5 profile"
echo "- Rust uses: cargo with targets x86_64-pc-windows-gnu and i686-pc-windows-gnu"
echo "You can run Templator by installing golang and running the following command: go build main.go to compile the project"
echo "You can also run the following command: go run main.go to run the project"

