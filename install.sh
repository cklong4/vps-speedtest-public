#!/bin/sh
# 公共一键入口：目标不需要Python/Go；只安装下载和证书基础工具。
set -eu
REPO=cklong4/vps-speedtest-public
VERSION=v3.0.1
case "$(uname -s)" in Linux) ;; *) printf '%s\n' '仅支持Linux VPS';exit 2;; esac
case "$(uname -m)" in x86_64|amd64) ARCH=amd64;; aarch64|arm64) ARCH=arm64;; *) printf '%s\n' '不支持的CPU架构；当前提供amd64/arm64';exit 2;; esac
have(){ command -v "$1" >/dev/null 2>&1; }
install_tools(){
    if [ "$(id -u)" != 0 ]; then printf '%s\n' '缺下载/校验工具；请root运行以自动安装curl和ca-certificates';exit 2;fi
    printf '%s\n' '安装最少基础依赖：下载工具和CA证书；不升级系统，不安装Python。'
    if have apt-get; then apt-get -o Acquire::ForceIPv4=true update && DEBIAN_FRONTEND=noninteractive apt-get -o Acquire::ForceIPv4=true install -y --no-install-recommends curl ca-certificates coreutils
    elif have apk; then apk add --no-cache curl ca-certificates coreutils
    elif have dnf; then dnf install -y curl ca-certificates coreutils
    elif have yum; then yum install -y curl ca-certificates coreutils
    else printf '%s\n' '没有支持的包管理器，请自行安装curl/wget、CA证书、sha256sum';exit 2;fi
}
if { ! have curl && ! have wget; } || ! have sha256sum || [ ! -r /etc/ssl/certs/ca-certificates.crt ]; then install_tools;fi
fetch(){
    if have curl;then curl -q --noproxy '*' --fail --silent --show-error --location --retry 2 --connect-timeout 15 --max-time 180 "$1" -o "$2"
    else env -u http_proxy -u https_proxy -u HTTP_PROXY -u HTTPS_PROXY -u ALL_PROXY wget -q -T 180 -O "$2" "$1";fi
}
DIR=$(mktemp -d "${TMPDIR:-/tmp}/vps-netcheck.XXXXXXXX")
trap 'rm -rf "$DIR"' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
FILE=vps-netcheck-linux-$ARCH
BASE=https://github.com/$REPO/releases/download/$VERSION
printf '%s\n' "下载 $VERSION 静态程序（$ARCH），无需预装Python3。"
# 发布包和校验表同样经过正常HTTPS，不关闭证书验证。
if ! fetch "$BASE/SHA256SUMS" "$DIR/SHA256SUMS";then
    printf '%s\n' '发布下载失败，尝试安装/更新最少CA证书工具后重试一次。'
    install_tools
    fetch "$BASE/SHA256SUMS" "$DIR/SHA256SUMS"
fi
fetch "$BASE/$FILE" "$DIR/$FILE"
EXPECTED=
while read -r HASH NAME;do
    if [ "$NAME" = "$FILE" ];then EXPECTED=$HASH;fi
done < "$DIR/SHA256SUMS"
case "$EXPECTED" in ''|*[!0-9a-f]*) printf '%s\n' '校验表无效，拒绝运行';exit 1;; esac
[ "${#EXPECTED}" = 64 ] || { printf '%s\n' '校验长度错误';exit 1; }
printf '%s  %s\n' "$EXPECTED" "$DIR/$FILE" | sha256sum -c -
chmod 700 "$DIR/$FILE"
"$DIR/$FILE" "$@"
