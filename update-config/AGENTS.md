# AGENTS.md

## 模块定位

`update-config/` 是订阅生成客户端脚本模块。GitHub Actions 只负责传入 secrets 和调用脚本；具体下载 subconverter、缓存外部规则、生成配置、加密和发起上传都在本模块维护。

服务端属于 `config-depot/`：服务端二进制、服务端数据目录、上传落点和缓存配置都集中在 `config-depot/`。

## 项目地图

| 路径 | 职责 |
|------|------|
| `install-dependencies.sh` | CI 安装 `jq`、`curl`、`tar`、`gzip`、`openssl` |
| `prepare.sh` | 把外部传入的密钥、上传地址和订阅内容写成流程输入文件 |
| `run-subconverter.sh` | 下载并启动 subconverter，调整 `base_path` |
| `cache-external-config.sh` | 下载 ACL4SSR，并把远程规则地址替换成本地缓存路径 |
| `update-config.sh` | 读取订阅列表，生成各 target 和 suffix 的配置文件 |
| `compress-config.sh` | 打包并按 config-depot 兼容格式加密 `config.tar.gz.aes` |
| `deploy-config.sh` | 调用 config-depot `/upload` 上传加密配置包 |
| `decrypt-config.sh` | artifact 路径下解密配置包，供 Actions 上传明文产物 |
| `local/` | 本地端到端验收脚本，每个脚本只做一个步骤 |
| `tmp/` | 客户端生成流程中间产物目录，由 `.gitignore` 忽略 |

## 本地验收

本地端到端入口：

```shell
update-config/local/e2e.sh
```

验收脚本默认使用 `config-depot/data` 中的旧版真实服务端数据，启动本地 config-depot，运行客户端生成、加密和上传流程，最终通过 `/upload` 覆盖 `config-depot/data/config`，然后检查 `/sub` 是否从服务端缓存读取。

需要访问 GitHub release API 时，可把临时 token 放在 `config-depot/data/github_token`。该文件在 `config-depot/.gitignore` 的 `/data/` 范围内，不应提交。

## 产物约束

- 客户端生成流程中间产物默认放在 `update-config/tmp/`。
- config-depot 的 Go 编译产物默认放在 `config-depot/config-depot`，使用 Go 项目已有忽略规则。
- 服务端数据不放入脚本模块；最终上传目标固定是 config-depot 数据目录，默认覆盖 `config-depot/data/config`。
- 如果临时改用 `WORK_DIR`、`CONFIG_DEPOT_BINARY` 等环境变量，目标路径必须是已有忽略规则覆盖的位置。
- 生成的密钥、订阅、压缩包、解密产物、PID、日志和下载包都不能进入 Git。
- macOS 本地打包必须保留 `COPYFILE_DISABLE=1`，避免 `._*` AppleDouble 文件进入上传包。

## 脚本边界

- CI workflow 不写业务逻辑，只调用本模块脚本。
- 外部信息只通过环境变量或输入文件传入脚本。
- 加密格式必须保持为：前 16 字节 IV，后续为 `openssl enc -aes-256-cbc -nosalt` 密文。
- 修改 `compress-config.sh` 或 config-depot 加解密逻辑后，必须做端到端上传验收。
