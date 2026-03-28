#!/bin/sh
set -e

# ============================================================
# frp-proxy release packaging script (Linux / macOS)
#
# Output layout:
#   release/frps            - server binary (direct-run / test)
#   release/frpc            - client binary (direct-run / test)
#   release/packages/
#     frps-<os>-<arch>-<ver>.tar.gz   (server package)
#     frpc-<os>-<arch>-<ver>.tar.gz   (client package)
#
# Server package contents:
#   frps-<os>-<arch>-<ver>/
#     frps
#     frps.toml
#     frps-service.sh    (Linux/macOS only)
#     LICENSE
#
# Client package contents:
#   frpc-<os>-<arch>-<ver>/
#     frpc
#     frpc.toml
#     LICENSE
# ============================================================

# 1. Build current platform binaries
make build

frp_version=$(./release/frps --version)
echo "build version: $frp_version"

# 2. Cross-compile all platforms
make -f ./Makefile.cross-compiles

# 3. Prepare output directory (keep historical packages, only clean current version)
mkdir -p ./release/packages

os_all='linux windows darwin freebsd openbsd'
arch_all='386 amd64 arm arm64 mips64 mips64le mips mipsle riscv64 loong64'

cd ./release

for os in $os_all; do
    for arch in $arch_all; do
        if [ "$os" = "windows" ]; then
            frps_bin="frps_${os}_${arch}.exe"
            frpc_bin="frpc_${os}_${arch}.exe"
        else
            frps_bin="frps_${os}_${arch}"
            frpc_bin="frpc_${os}_${arch}"
        fi

        # Skip if cross-compiled binaries don't exist
        [ -f "./$frps_bin" ] || continue
        [ -f "./$frpc_bin" ] || continue

        frps_dir="frps-${os}-${arch}-${frp_version}"
        frpc_dir="frpc-${os}-${arch}-${frp_version}"

        # --- Server package ---
        mkdir -p "./packages/$frps_dir"
        if [ "$os" = "windows" ]; then
            mv "./$frps_bin" "./packages/$frps_dir/frps.exe"
        else
            mv "./$frps_bin" "./packages/$frps_dir/frps"
            cp ../frps-service.sh "./packages/$frps_dir/frps-service.sh"
            chmod +x "./packages/$frps_dir/frps-service.sh"
        fi
        cp ../conf/frps.toml "./packages/$frps_dir/frps.toml"
        cp ../LICENSE "./packages/$frps_dir/LICENSE"

        cd ./packages
        if [ "$os" = "windows" ]; then
            zip -rq "$frps_dir.zip" "$frps_dir"
        else
            tar -zcf "$frps_dir.tar.gz" "$frps_dir"
        fi
        rm -rf "./$frps_dir"
        cd ..

        # --- Client package ---
        mkdir -p "./packages/$frpc_dir"
        if [ "$os" = "windows" ]; then
            mv "./$frpc_bin" "./packages/$frpc_dir/frpc.exe"
        else
            mv "./$frpc_bin" "./packages/$frpc_dir/frpc"
        fi
        cp ../conf/frpc.toml "./packages/$frpc_dir/frpc.toml"
        cp ../LICENSE "./packages/$frpc_dir/LICENSE"

        cd ./packages
        if [ "$os" = "windows" ]; then
            zip -rq "$frpc_dir.zip" "$frpc_dir"
        else
            tar -zcf "$frpc_dir.tar.gz" "$frpc_dir"
        fi
        rm -rf "./$frpc_dir"
        cd ..
    done
done

cd -
echo "Done. Packages are in release/packages/"
