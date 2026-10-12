VPS 一键网络测试（公共版 v3.1.0）
================================
仓库：https://github.com/cklong4/vps-speedtest-public

默认流程：先测速 → 再探测去程/回程线路参考 → 四来源10分钟延迟与波动 → 一份完整报告和7天网页链接。
目标VPS无需预装Python3或Go，无需Token、GitHub Key、来源SSH密钥。Linux amd64/arm64静态程序，SHA256校验后运行。

一行命令
--------
短入口（已有curl）：
curl -fsSL https://nezha.hzdbg.com/vps.sh | sh
没有curl但有wget：
wget -qO- https://nezha.hzdbg.com/vps.sh | sh
这两个短命令同样自动执行先速度、再线路、四来源与7天报告；依赖检测留在脚本内部。curl和wget都没有的极简系统，仍需先安装下载工具，或使用下面完整自动引导的一行命令。

GitHub源码入口（已有curl）：
curl -fsSL --max-time 20 https://raw.githubusercontent.com/cklong4/vps-speedtest-public/main/install.sh | sh
GitHub raw不可达时：
curl -fsSL --max-time 30 https://nezha.hzdbg.com/vps-download/v3.1.0/install.sh | sh

刚重装、连curl都没有：复制“一行命令.txt”的整行。该命令自动按apt/apk/dnf/yum安装curl和CA证书再运行，不安装Python，不现场编译。安装缺失依赖需要root。包管理器网络或GitHub/备用HTTPS均不可达时如实报错，不关闭证书验证。
也可先保存、审阅install.sh再执行。参数：sh install.sh --port 443，SSH或节点端口不是22时必须指定正确公开端口；--output指定结果目录；--local-only仅本机速度/回程参考；--resume <本轮JSON>只查看既有任务，不重复开测。

完整报告
--------
1. 本机速度：Cloudflare HTTPS单连接/4连接下载与上传，约60MiB载荷（不含协议和重传）。有限载荷端到端平均速度，不是持续带宽上限，不是国内家宽至VPS的速度。速度数据由客户端自己测量，报告注明并非协调端独立验证。
2. 回程参考：VPS到电信202.101.172.35（AS4134）、联通210.22.70.3（AS17621）、移动211.136.192.6（AS56040）的UDP TTL路径；每目标最多24跳，有界等待。不是覆盖三网全国所有地区。没有raw socket权限时注明未测，不自动提权/改防火墙。
3. 去程：杭州家宽、杭州阿里云、广州火山云、香港RFC实际来源到VPS的TCP MTR，分别保留跳点/缺口及ASN依据。
4. 线路识别：起源ASN来自RIPEstat network-info；提供电信163、CN2特征、CMI、联通169/精品网络等提示。不把少数59.43/AS4809跳点当CN2 GIA认证；混合163/CN2路径也不能唯一认证GT。无法确定商业等级就写无法确认。
5. 延迟/波动：四来源共享正式10分钟窗口；Ping真实收发、中位/P95/范围、相邻成功包波动；实际TCP端口建连成功/失败、P95和尖峰。ICMP与TCP MTR各60轮短旁证，终点不足60次时保留不完整，不冒称0丢包。
6. 本机带宽先完成才运行延迟采样，不用满速下载干扰Ping。已结束但有网络/采样缺口会生成报告并返回非零，不能把整个缺口标成全部成功。

公开匿名接口
------------
服务地址：https://nezha.hzdbg.com/vps-api/
只测实际连接来源自身的公网IPv4；不能指定别人IP，X-Real-IP由可信反代真实remote_addr覆盖，客户端XFF不可信。
额度：每IP每小时1次、24小时3次；同/24网段24小时6次；全局24小时24次；一次只跑一轮，忙时明确拒绝不排多轮。额度跨服务重启保留。
匿名流程完全不调用私有节点UUID、密码或代理配置。公共四来源测的是Ping/TCP/MTR，不是你的VLESS/SS完整代理HTTPS。原私库和已授权私有版流程保留。
客户端产生的本轮查看JSON只是随机任务查看凭据，无需你预先申请或填写Key，权限0600。不要公开该JSON；报告共享URL只看报告，不授任务管理权限。

七天链接
--------
结束后显示：完整本地TXT、免登录网页链接、准确到期时间。链接从生成的finished时间起7天，读取不续期，到期410、错误串404。页面可下载TXT，移动端可直接读。
拿到链接的人都能看公网IP和路由信息，勿公开转发敏感报告。报告不会列出来源SSH凭据，不自动推送聊天。到期只是公开访问失效；本地TXT和协调端私密原始证据不自动删。

下载与兼容性
------------
GitHub公开发布包含amd64和arm64静态程序、SHA256SUMS、源码。GitHub Release CDN不可达时转同版本HTTPS镜像，仍校验SHA256；不使用-k。
支持常见Debian/Ubuntu、Alpine、RHEL系Linux。已实测无Python/Go的Alpine本机测速和无Python/curl/wget的Ubuntu24基础依赖安装→静态运行。Debian本机测试曾因deb.debian.org网络超时未通过，网络不可达不保证可安装；arm64已构建并校验，未在arm64硬件实测。
安装只补最少依赖，可能连带更新其必需系统库；不执行全系统upgrade、不改BBR/SSH/代理/防火墙/路由、不装常驻客户端服务。

开发验证
--------
GO111MODULE=off go test -race ./client
python3 -m unittest discover -s tests -v
构建：CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off go build -trimpath -ldflags '-s -w' -o release/vps-netcheck-linux-amd64 ./client
校验表必须使用文件basename，不得带release/目录前缀。静态程序的目标机运行不依赖任何Python测试工具。
服务端控制代码与私密来源接入留在原私库，公共仓库不发布历史测量、IP清单或凭据。公开仓库只发布干净公共客户端和启动器。
