# AGENTS.md

## 项目定位

SubConfig 的核心产物是 `subconverter*.ini` 及其引用的规则文件。任何 subconverter 实例都可以通过 `config` 参数直接引用这组配置，无需本仓库的其他部分参与。

在此基础上，仓库用 GitHub Actions 按这组配置转换作者自己的机场订阅，并由独立的 Go 服务 `config-depot/` 接收加密配置包、提供下载。自动化流程是这组配置的一种使用方式，不是配置生效的前提。

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
| `update-config/` | 订阅生成客户端脚本模块，详见 `update-config/AGENTS.md` |
| `config-depot/` | Go 发布服务，详见 `config-depot/AGENTS.md` |
| `tmp/` | 项目内临时目录，内容由 `tmp/.gitignore` 忽略 |

## 工作边界

- 规则必须自包含在外部配置中，不得改为依赖 subconverter 部署环境的 `pref.ini` 或其他全局设置，否则外部引用者拿不到完整行为。
- `exclude_remarks` 不支持 `!!import`，只能在每个配置变体中直接写正则；同一条排除规则在多个文件中重复维护是上述两项约束的结果，不是待清理的冗余。
- 根目录配置文件和 `subconverter.yml` 属于订阅生成流程。
- 订阅生成客户端脚本放在 `update-config/`，workflow 只调用脚本，不内联业务逻辑。
- `config-depot/` 属于发布服务，和订阅生成 workflow 分开维护。
- 不要把 Go 测试或 Docker 构建加回 `.github/workflows/subconverter.yml`；Go 项目验证放在 `.github/workflows/config-depot.yml`。
- 发布服务的公开入口是 `/upload`、`/sub` 和 `/healthz`。
- 临时演练产物放仓库 `tmp/`，不要写系统临时目录。

## 常用验证

订阅更新 workflow 依赖外部服务和 GitHub secrets，本地通常只检查 YAML 改动和相关 shell 逻辑。`config-depot` 的验证命令见 `config-depot/AGENTS.md`。
