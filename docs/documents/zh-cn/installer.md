# ToughRADIUS 快速安装指南

ToughRADIUS 是一种健壮、高性能、易于扩展的开源 RADIUS 服务器。本指南将引导您快速地在您的系统上安装和配置 ToughRADIUS 服务。

## 快速安装

### 通过 curl 或 wget 安装

您可以使用 `curl` 或 `wget` 工具来快速安装 ToughRADIUS。根据您的喜好选择以下命令之一执行即可。

使用 curl 安装：

  ```bash
  sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/talkincode/toughradius/main/installer.sh)"
  ```
  
使用 wget 安装：

  ```bash
  sudo bash -c "$(wget https://raw.githubusercontent.com/talkincode/toughradius/main/installer.sh -O -)"
  ```

### 二进制安装

我们以 v8.0.6 版本为例进行安装说明。请依据以下步骤进行：

- 从 [Releases 页面](https://github.com/talkincode/toughradius/releases) 下载软件发行版。

- 如果您具备一定的开发能力，您也可以选择自行编译版本。

使用 `curl` 下载 ToughRADIUS 并进行安装：

  ```bash
  curl https://github.com/talkincode/toughradius/releases/download/v8.0.6/toughradius_amd64 -O /tmp/toughradius

  chmod +x /tmp/toughradius && /tmp/toughradius -install
  ```

### 系统环境依赖

在开始安装之前，请确保您的系统满足以下条件：

- 操作系统：支持跨平台部署（Linux、Windows、MacOS 等）
- 数据库服务器：PostgreSQL 14 或更高版本

### 数据库初始化

在进行 ToughRADIUS 的安装和配置之前，请确保您的数据库服务器已经正确安装并且正在运行。下面是数据库的初始化步骤：

- 运行数据库创建脚本并创建一个专用用户：

  ```sql
  CREATE USER toughradius WITH PASSWORD 'toughradius';
  CREATE DATABASE toughradius WITH OWNER toughradius;
  GRANT ALL PRIVILEGES ON DATABASE toughradius TO toughradius;
  ```

请将 `toughradius` 替换成您想要设置的密码。

- 在继续操作之前，请确保您已经创建了相应的数据库，并确保数据库服务器正在运行。

- 修改配置文件 `/etc/toughradius.yml`。

### 启动服务

安装完成后，您可以通过以下命令启动 ToughRADIUS 服务，并设置为开机自启：

```bash
systemctl enable toughradius
systemctl start toughradius
```

### 访问控制台

打开您的网络浏览器，输入 URL：`http://服务器IP:1816`。请将 "服务器IP" 替换成您的服务器实际的 IP 地址。

默认的用户名和密码是：`admin/toughradius`

至此，ToughRADIUS 的安装和基本配置已经完成。您现在可以开始配置您的 RADIUS 服务器，并管理您的用户认证和账户计费。