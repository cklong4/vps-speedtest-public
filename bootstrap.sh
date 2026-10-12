#!/bin/sh
# 先补下载工具，再取得真正启动器；即使没有Python/Go/curl/wget也支持常见root新装Linux。
set -eu
have(){ command -v "$1" >/dev/null 2>&1; }
if { ! have curl && ! have wget; } || [ ! -r /etc/ssl/certs/ca-certificates.crt ];then
 [ "$(id -u)" = 0 ] || { printf '%s\n' '需要root安装最少下载工具/CA证书';exit 2; }
 if have apt-get;then apt-get -o Acquire::ForceIPv4=true update && apt-get -o Acquire::ForceIPv4=true install -y --no-install-recommends curl ca-certificates
 elif have apk;then apk add --no-cache curl ca-certificates
 elif have dnf;then dnf install -y curl ca-certificates
 elif have yum;then yum install -y curl ca-certificates
 else printf '%s\n' '找不到支持的包管理器';exit 2;fi
fi
D=$(mktemp -d "${TMPDIR:-/tmp}/vps-bootstrap.XXXXXXXX")
trap 'rm -rf "$D"' EXIT
URL=https://raw.githubusercontent.com/cklong4/vps-speedtest-public/main/install.sh
FALLBACK=https://nezha.hzdbg.com/vps-download/v3.1.0/install.sh
if have curl;then
 curl -q --noproxy '*' -fSsL --max-time 20 "$URL" -o "$D/install.sh" || curl -q --noproxy '*' -fSsL --max-time 30 "$FALLBACK" -o "$D/install.sh"
else
 wget -q -T 20 -O "$D/install.sh" "$URL" || wget -q -T 30 -O "$D/install.sh" "$FALLBACK"
fi
sh "$D/install.sh" "$@"
