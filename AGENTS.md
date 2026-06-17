# AGENTS.md

## 项目定位

SubConfig 是一组 subconverter 外部配置，以及配套的 GitHub Actions 自动更新流程。仓库根目录主要维护静态配置文件；`config-depot/` 是独立的 Go 服务，用来接收 Actions 生成的加密配置包并提供下载。

## 项目地图

| 路径 | 职责 |
|------|------|
| `subconverter*.ini` | subconverter 使用的主配置，后缀代表不同配置变体 |
| `base_config.yml`、`base_quan.conf`、`base_singbox.json` | 不同客户端的基础配置模板 |
| `*.list`、`surge_ruleset*.txt`、`custom_proxy_group*.txt` | 分流规则、代理组和规则列表 |
| `exclude.ini`、`exclude.txt`、`rename.txt` | 订阅转换时使用的过滤和改名辅助配置 |
| `.github/workflows/subconverter.yml` | 自动运行 subconverter，生成、加密、上传订阅配置 |
| `.github/workflows/config-depot.yml` | 只验证 `config-depot` Go 项目 |
| `.github/workflows/debugger.yml` | 手动调试订阅更新流程 |
| `config-depot/` | Go 发布服务，详见 `config-depot/AGENTS.md` |
| `tmp/` | 项目内临时目录，内容由 `tmp/.gitignore` 忽略 |

## 工作边界

- 根目录配置文件和 `subconverter.yml` 属于订阅生成流程。
- `config-depot/` 属于发布服务，和订阅生成 workflow 分开维护。
- 不要把 Go 测试或 Docker 构建加回 `.github/workflows/subconverter.yml`；Go 项目验证放在 `.github/workflows/config-depot.yml`。
- 发布服务的公开入口是 `/upload`、`/sub` 和 `/healthz`。
- 临时演练产物放仓库 `tmp/`，不要写系统临时目录。

## 常用验证

订阅更新 workflow 依赖外部服务和 GitHub secrets，本地通常只检查 YAML 改动和相关 shell 逻辑。`config-depot` 的验证命令见 `config-depot/AGENTS.md`。
