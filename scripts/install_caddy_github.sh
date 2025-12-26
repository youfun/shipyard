#!/bin/bash

# ==============================================================================
# 脚本名称: install_caddy_universal.sh
# 描述:     在主流 Linux 系统上从 GitHub Releases 下载并安装最新版 Caddy。
#           此脚本会自动检测包管理器 (APT, YUM/DNF, Pacman) 并安装所需依赖。
# 作者:     Gemini
# 日期:     2025-08-30
# ==============================================================================

# 在执行命令时输出，并在出错时立即退出
set -e
set -o pipefail

# --- 权限检查 ---
if [ "$(id -u)" -ne 0 ]; then
  echo "错误：此脚本必须以 root 权限运行。" >&2
  echo "请尝试使用 'sudo ./install_caddy_universal.sh' 命令运行。" >&2
  exit 1
fi

# --- 依赖项安装 ---
echo ">>> 步骤 1: 检测包管理器并安装依赖项 (curl)..."
PKG_MANAGER=""
INSTALL_CMD=""

if command -v apt-get &> /dev/null; then
  PKG_MANAGER="apt"
  INSTALL_CMD="apt-get install -y"
  # 在 Debian/Ubuntu 上，先更新包列表
  echo "检测到 Debian/Ubuntu (apt)。正在更新包列表..."
  apt-get update
elif command -v dnf &> /dev/null; then
  PKG_MANAGER="dnf"
  INSTALL_CMD="dnf install -y"
  echo "检测到 Fedora/RHEL/AlmaLinux/Rocky Linux (dnf)..."
  # 检查具体发行版
  if [ -f /etc/os-release ]; then
    source /etc/os-release
    echo "检测到发行版: $NAME $VERSION_ID"
  fi
elif command -v yum &> /dev/null; then
  PKG_MANAGER="yum"
  INSTALL_CMD="yum install -y"
  echo "检测到 CentOS/RHEL (yum)..."
elif command -v pacman &> /dev/null; then
  PKG_MANAGER="pacman"
  # Pacman 需要非交互式参数
  INSTALL_CMD="pacman -S --noconfirm"
  echo "检测到 Arch Linux (pacman)..."
else
  echo "错误：无法识别的包管理器。请手动安装 'curl' 后再运行此脚本。" >&2
  exit 1
fi

# 检查并安装 curl
if ! command -v curl &> /dev/null; then
  echo "正在安装 'curl'..."
  ${INSTALL_CMD} curl
else
  echo "'curl' 已安装。"
fi

# --- 自动检测架构和构建下载链接 ---
echo ""
echo ">>> 步骤 2: 检测系统架构..."

case "$(uname -m)" in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  armv6l | armv7l) ARCH="armv7" ;;
  *)
    echo "错误：不支持的系统架构: $(uname -m)" >&2
    exit 1
    ;;
esac
echo "检测到系统架构: $ARCH"

echo ">>> 构建下载链接..."
DOWNLOAD_URL="https://github.com/openmindw/caddy-cloudflare/releases/download/latest/caddy-cloudflare-linux-${ARCH}.tar.gz"
echo "下载链接: ${DOWNLOAD_URL}"

# --- 下载和安装 ---
echo ""
echo ">>> 步骤 3: 下载并安装 Caddy (Cloudflare) 二进制文件..."
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

curl -sSL "$DOWNLOAD_URL" -o "caddy.tar.gz"

# 先列出压缩包内容以便调试
echo "压缩包内容："
tar -tzf "caddy.tar.gz"

# 解压所有内容
tar -xzf "caddy.tar.gz"

# 查找 caddy 可执行文件
CADDY_BINARY=""
if [ -f "caddy" ]; then
    CADDY_BINARY="caddy"
elif [ -f "caddy-linux-${ARCH}" ]; then
    CADDY_BINARY="caddy-linux-${ARCH}"
elif [ -f "caddy-cloudflare" ]; then
    CADDY_BINARY="caddy-cloudflare"
else
    # 查找任何可执行文件
    CADDY_BINARY=$(find . -type f -executable -name "*caddy*" | head -1)
    if [ -z "$CADDY_BINARY" ]; then
        echo "错误：在压缩包中找不到 caddy 可执行文件。" >&2
        echo "压缩包内容："
        ls -la
        exit 1
    fi
    CADDY_BINARY=$(basename "$CADDY_BINARY")
fi

echo "找到 Caddy 可执行文件: $CADDY_BINARY"
echo "将 Caddy 可执行文件移动到 /usr/local/bin/..."
mv "$CADDY_BINARY" /usr/local/bin/caddy
chown root:root /usr/local/bin/caddy
chmod +x /usr/local/bin/caddy

cd ..
rm -rf "$TMP_DIR"

# --- 创建用户和目录 ---
echo ""
echo ">>> 步骤 4: 创建 Caddy 用户、组和所需目录..."
if ! getent group caddy > /dev/null; then
  echo "创建 'caddy' 组..."
  groupadd --system caddy
fi
if ! id caddy > /dev/null 2>&1; then
  echo "创建 'caddy' 用户..."
  useradd --system --gid caddy  --shell /usr/sbin/nologin caddy
fi

mkdir -p /etc/caddy
chown -R root:caddy /etc/caddy
mkdir -p /var/lib/caddy
chown -R caddy:caddy /var/lib/caddy


# 新增：创建配置目录并修复权限
mkdir -p /home/caddy/.config/caddy
chown -R caddy:caddy /home/caddy


# --- 设置 systemd 服务 ---
echo ""
echo ">>> 步骤 5: 设置 Caddy 的 systemd 服务..."
cat <<EOF > /etc/systemd/system/caddy.service
[Unit]
Description=Caddy
Documentation=https://caddyserver.com/docs/
After=network.target network-online.target
Requires=network-online.target

[Service]
Type=notify
User=caddy
Group=caddy
ExecStart=/usr/local/bin/caddy run --environ --config /etc/caddy/caddy.json --resume
ExecReload=/usr/local/bin/caddy reload --config /etc/caddy/caddy.json --resume
TimeoutStopSec=5s
LimitNOFILE=1048576
LimitNPROC=512
PrivateTmp=true
ProtectSystem=full
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF

# --- 创建默认配置文件 ---
echo ""
echo ">>> 步骤 6: 创建一个默认的 caddy.json..."
if [ ! -f /etc/caddy/caddy.json ]; then
cat <<EOF > /etc/caddy/caddy.json
{
  "apps": {
    "http": {
      "servers": {
        "srv0": {
          "listen": [
            ":80",
            ":443"
          ],
          "protocols": [
            "h1",
            "h2"
          ],
          "routes": []
        }
      }
    },
    "tls": {
      "automation": {}
    }
  }
}
EOF
chown caddy:caddy /etc/caddy/caddy.json
chmod 644 /etc/caddy/caddy.json
fi

# --- 启动服务 ---
echo ""
echo ">>> 步骤 7: 重新加载 systemd 并启动 Caddy 服务..."
systemctl daemon-reload
systemctl enable --now caddy

# --- 验证 ---
echo ""
echo ">>> 步骤 8: 验证安装..."
sleep 2 # 稍等片刻以确保服务完全启动

if ! command -v caddy &> /dev/null; then
    echo "错误：Caddy 安装失败，找不到命令。" >&2
    exit 1
fi

echo "Caddy 安装成功！"
caddy version

echo ""
echo "Caddy 服务状态："
systemctl status caddy --no-pager | cat # 使用 cat 防止在脚本中分页

echo ""
echo -e "\033[32m安装完成！\033[0m"
echo "您可以通过访问 http://<您的服务器IP> 来测试默认页面。"
echo "配置文件位于 /etc/caddy/Caddyfile。"
echo "修改配置后，请运行 'sudo systemctl reload caddy' 应用更改。"

exit 0