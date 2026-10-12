v3.0.0 纯一键 VPS 网络测试
===========================

公共仓库将提供：
- 先测本机下载/上传，再测线路参考和四来源延迟/波动。
- 目标机不需要 Python、Go 或预装测速软件。
- Shell 启动器自动识别 amd64/arm64，下载静态程序并校验 SHA256。
- 缺 curl、CA证书、sha256sum 时，在 root 权限下只安装最少基础依赖，不升级整个系统。
- 完成后生成速度、ASN/线路证据、Ping、TCP、MTR和稳定性报告，并提供7天免登录报告链接。

一行运行
--------
```sh
curl -fsSL https://raw.githubusercontent.com/cklong4/vps-speedtest-public/main/install.sh | sh
```

这条命令只会下载并校验公开发布的静态程序。若系统缺 curl，需先用系统包管理器安装 curl；启动器自身只能在已有下载工具时运行。若使用 root，缺 curl/CA/sha256sum 时启动器会安装最少依赖：curl、ca-certificates、coreutils；不会安装 Python。

也可以先下载后审阅：
```sh
curl -fsSLo install.sh https://raw.githubusercontent.com/cklong4/vps-speedtest-public/main/install.sh
less install.sh
sh install.sh
```

运行范围
--------
默认流程固定为：
1. 本机到 Cloudflare HTTPS 的单连接/4连接下载和上传，计划载荷约60MiB。
2. VPS到电信、联通、移动公网参考目标的UDP TTL线路参考；不保证所有路由器回应。
3. 通过公开协调端，从杭州家宽、杭州阿里云、广州火山云、香港RFC测10分钟 TCP/ICMP/MTR稳定性。
4. 把本机速度、线路特征和四来源结果合并为一份报告。

匿名接口安全边界
----------------
公共接口不接受Token或GitHub Key，但只允许测试“当前连接来源自己的公网IPv4”：
- 经可信反代只采信代理写入的真实 remote_addr；客户端的 X-Forwarded-For/X-Real-IP 不能让调用者伪造测试目标。
- 每个公网IP每小时1次、每天3次；同一/24网段每天6次；全局每天24次；一次只跑一轮。
- 匿名任务不调用、不接触私有节点账号、UUID、SSH密钥或密码；只测Ping/TCP/MTR。
- 报告中的公网IP、跳点和线路信息可能泄露网络信息；拿到7天链接的人都能查看。
- 公开测试是有限载荷、有限时长诊断，不是开放测速站或带宽压测服务。

线路识别边界
------------
报告将列出观测方向、MTR跳点IP、RIPEstat返回的起源ASN和IP前缀，并给出“163 / CN2特征 / CMI / 联通特征 / 未知”等结果。
- AS4134及202.97段通常只能说明163路径特征。
- AS4809或59.43跳点只能说明CN2特征，不能单凭少数跳点认证CN2 GIA。
- CN2 GT/GIA、去程/回程、商业产品等级必须分开，证据不足明确写“无法确认”。
- CMI等ASN是观测网络身份，不等于所有路径都经过相同移动国际出口。
- 星号、隐藏跳点、路由不对称、ICMP限速不能当业务丢包或运营商责任证明。

报告链接
--------
测试结束后输出类似：
`https://nezha.hzdbg.com/vps-report/<随机串>`

链接从本轮结束时间起7天有效，免登录，可打开网页或追加`.txt`下载。到期自动返回410，不会续期。旧私有仓库和私有版Token流程保留，公共版不包含节点配置、凭据或用户历史报告。

参数
----
```sh
sh install.sh --version
sh install.sh --local-only       # 只测速度和线路参考，不调用四来源
sh install.sh --port 443         # 目标VPS不是22时指定节点/SSH端口
sh install.sh --output /root/vps-report
```

安全说明
--------
- 安装器只支持Linux amd64/arm64；静态程序通过SHA256SUMS校验后才运行。
- 下载使用正常HTTPS证书校验，不使用 `-k`、不把Token嵌入URL。
- 程序不修改BBR、防火墙、代理、SSH或系统路由；只生成临时文件和本地报告。
- 终端Ctrl+C会停止本机阶段；已创建的四来源任务按10分钟截止，不会无限运行。重复任务会受到服务端全局忙和IP配额限制。
- 这不是全天稳定性保证。公网测速目标、Anycast、来源WAN和当前路由都会影响结果。
