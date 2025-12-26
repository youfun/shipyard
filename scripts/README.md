# Shipyard 安装脚本

本目录包含 Shipyard Server 和 CLI 的自动安装脚本。

## 📦 可用脚本

### 1. install-shipyard-cli.sh

自动检测系统并从 GitHub Release 安装 shipyard-cli 客户端。

**特性：**
- ✅ 自动检测操作系统（Linux/macOS）
- ✅ 自动检测架构（amd64/arm64）
- ✅ 从 GitHub Release 下载最新版本或指定版本
- ✅ 智能选择安装目录（root 用户安装到 `/usr/local/bin`，普通用户安装到 `~/.local/bin`）
- ✅ 自动添加执行权限
- ✅ 验证安装并提供使用指南

**使用方法：**

```bash
# 一键安装最新版本（推荐）
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/deployer/main/scripts/install-shipyard-cli.sh | bash

# 安装指定版本
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/deployer/main/scripts/install-shipyard-cli.sh | bash -s v1.0.0

# 或者下载后本地执行
wget https://raw.githubusercontent.com/YOUR_ORG/deployer/main/scripts/install-shipyard-cli.sh
chmod +x install-shipyard-cli.sh
./install-shipyard-cli.sh           # 安装最新版本
./install-shipyard-cli.sh v1.0.0    # 安装指定版本
```

**安装位置：**
- Root 用户: `/usr/local/bin/shipyard-cli`
- 普通用户: `~/.local/bin/shipyard-cli`

**首次使用：**

```bash
# 查看版本
shipyard-cli --version

# 登录到服务器
shipyard-cli login --endpoint http://your-server:8080

# 测试连接
shipyard-cli ping
```

---

### 2. install-shipyard-server.sh

从 GitHub Release 下载并安装 shipyard-server 到系统，支持标准模式和测试模式。

**特性：**
- ✅ 自动检测系统架构（amd64/arm64）
- ✅ 从 GitHub Release 下载最新版本或指定版本
- ✅ 两种安装模式：
  - **标准模式**：安装到系统目录，创建 systemd 服务
  - **测试模式**：安装到当前目录，无需 root 权限
- ✅ 自动创建系统用户和组（标准模式）
- ✅ 生成环境配置文件模板
- ✅ 配置 systemd 服务（标准模式）

**使用方法：**

```bash
# 一键安装最新版本（标准模式，需要 sudo）
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/deployer/main/scripts/install-shipyard-server.sh | sudo bash

# 安装指定版本
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/deployer/main/scripts/install-shipyard-server.sh | sudo bash -s v1.0.0

# 或者下载后本地执行
wget https://raw.githubusercontent.com/YOUR_ORG/deployer/main/scripts/install-shipyard-server.sh
chmod +x install-shipyard-server.sh
sudo ./install-shipyard-server.sh           # 安装最新版本
sudo ./install-shipyard-server.sh v1.0.0    # 安装指定版本
```

**安装模式选择：**

运行脚本后会提示选择：
1. **标准模式** - 安装到系统目录（需要 root 权限）
   - 二进制文件: `/usr/local/bin/shipyard-server`
   - 配置目录: `/etc/shipyard`
   - 数据目录: `/var/lib/shipyard`
   - 日志目录: `/var/log/shipyard`
   - systemd 服务: `shipyard-server.service`

2. **测试模式** - 安装到当前目录（无需 root 权限）
   - 所有文件在当前目录下
   - 不创建 systemd 服务
   - 适合测试和开发

**配置和启动（标准模式）：**

```bash
# 1. 编辑配置文件（必须）
sudo nano /etc/shipyard/.env

# 2. 生成 JWT Secret
openssl rand -base64 32

# 3. 启动服务
sudo systemctl start shipyard-server

# 4. 查看状态
sudo systemctl status shipyard-server

# 5. 查看日志
sudo journalctl -u shipyard-server -f

# 6. 设置开机自启（可选）
sudo systemctl enable shipyard-server
```

**配置和启动（测试模式）：**

```bash
# 1. 编辑配置文件
nano config/.env

# 2. 启动服务
export $(cat config/.env | xargs)
./shipyard-server-linux-amd64 --port 8080
```

---

## 🔧 配置文件说明

安装后会自动生成 `.env` 配置文件，主要配置项：

```bash
# JWT 密钥（必须修改！）
JWT_SECRET=PLEASE_CHANGE_THIS_TO_RANDOM_SECRET

# 数据库类型
DB_TYPE=sqlite

# SQLite 数据库路径
DB_PATH=/var/lib/shipyard/deploy.db

# 服务器端口
SERVER_PORT=8080

# 日志级别
LOG_LEVEL=info
```

**重要提醒：** 请务必修改 `JWT_SECRET` 为随机密钥！

---

## 📝 更新脚本中的仓库地址

在使用前，请修改脚本中的 `GITHUB_REPO` 变量为你的实际仓库地址：

```bash
# 将此行
GITHUB_REPO="YOUR_ORG/deployer"

# 修改为
GITHUB_REPO="yourusername/deployer"
```

或者在 GitHub Release 时提供正确的下载链接。

---

## 🐛 故障排除

### 下载失败

```bash
# 检查网络连接
curl -I https://github.com

# 手动下载
wget https://github.com/YOUR_ORG/deployer/releases/download/v1.0.0/shipyard-cli-linux-amd64
chmod +x shipyard-cli-linux-amd64
sudo mv shipyard-cli-linux-amd64 /usr/local/bin/shipyard-cli
```

### PATH 问题

如果安装到 `~/.local/bin` 后无法使用命令：

```bash
# 添加到 PATH（bash）
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc
source ~/.bashrc

# 添加到 PATH（zsh）
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.zshrc
source ~/.zshrc
```

### 权限问题

```bash
# 标准模式需要 sudo
sudo ./install-shipyard-server.sh

# 或使用测试模式（无需 sudo）
./install-shipyard-server.sh  # 选择选项 2（测试模式）
```

---

## 📚 更多信息

- [主 README](../README.md)
- [GitHub Releases](https://github.com/YOUR_ORG/deployer/releases)
- [问题反馈](https://github.com/YOUR_ORG/deployer/issues)
