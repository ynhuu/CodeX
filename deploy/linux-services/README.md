# Linux Service Files

这个目录提供常见 Linux 发行版的服务文件示例。

- `systemd/`：适用于 Debian、Ubuntu 以及大多数使用 systemd 的发行版
- `openrc/`：适用于 Alpine Linux

## 推荐目录

- 程序：`/opt/codex-web/codex-web`
- 配置：`/etc/codex-web/conf.yml`
- 环境变量：`/etc/default/codex-web` 或 `/etc/conf.d/codex-web`

## Debian / Ubuntu

复制服务文件：

```bash
sudo cp deploy/linux-services/systemd/codex-web.service /etc/systemd/system/
sudo cp deploy/linux-services/systemd/codex-web.env.example /etc/default/codex-web
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now codex-web
sudo systemctl status codex-web
```

## Alpine

复制服务文件：

```bash
sudo cp deploy/linux-services/openrc/codex-web /etc/init.d/codex-web
sudo chmod +x /etc/init.d/codex-web
sudo cp deploy/linux-services/openrc/codex-web.confd /etc/conf.d/codex-web
```

启动服务：

```bash
sudo rc-update add codex-web default
sudo rc-service codex-web start
sudo rc-service codex-web status
```
