# PicoClaw 服务管理指南

## launchd 服务管理（macOS）

PicoClaw gateway 可以配置为 macOS 后台服务，开机自动启动。

### 安装服务

```bash
# 复制服务文件
cp /Volumes/ORICO/workspace/github/picoclaw/com.picoclaw.gateway.plist ~/Library/LaunchAgents/

# 加载并启动服务
launchctl load ~/Library/LaunchAgents/com.picoclaw.gateway.plist
```

### 常用管理命令

| 操作 | 命令 |
|------|------|
| 查看状态 | `launchctl list \| grep picoclaw` |
| 停止服务 | `launchctl stop com.picoclaw.gateway` |
| 启动服务 | `launchctl start com.picoclaw.gateway` |
| 卸载服务 | `launchctl unload ~/Library/LaunchAgents/com.picoclaw.gateway.plist` |
| 查看日志 | `tail -f ~/.picoclaw/gateway.log` |

### 服务配置说明

- **日志文件**: `~/.picoclaw/gateway.log`
- **配置文件**: `~/.picoclaw/config.json`
- **开机自启**: 是（RunAtLoad + KeepAlive）

---

## 快速命令行操作

```bash
# 直接启动（前台运行，Ctrl+C 停止）
picoclaw gateway

# 后台运行（nohup 方式）
nohup picoclaw gateway > ~/.picoclaw/gateway.log 2>&1 &

# 停止后台进程
pkill -f "picoclaw gateway"

# 查看进程
ps aux | grep picoclaw
```

---

## 故障排查

### 查看服务是否运行

```bash
launchctl list | grep picoclaw
```

如果看到类似输出，说明服务正在运行：
```
12345  0  com.picoclaw.gateway
```

### 查看日志

```bash
tail -100 ~/.picoclaw/gateway.log
```

### 重启服务

```bash
launchctl stop com.picoclaw.gateway
launchctl start com.picoclaw.gateway
```

### 代理问题

如果在国内，确保配置文件中 Telegram 代理已设置：

```json
{
  "channels": {
    "telegram": {
      "proxy": "http://127.0.0.1:7890"
    }
  }
}
```
