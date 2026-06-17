# AGENTS.md

## 项目定位

`config-depot` 是 SubConfig 的 Go 发布服务。它接收 GitHub Actions 上传的 `config.tar.gz.aes`，验证上传 token，解密并解包到数据目录的 `config/`，再通过 `/sub` 提供缓存配置下载。

## 项目地图

| 路径 | 职责 |
|------|------|
| `cmd/config-depot/main.go` | 进程入口，读取环境变量并启动 HTTP 服务 |
| `internal/app/defaults.go` | 默认值和环境变量名，修改默认配置优先看这里 |
| `internal/app/config.go` | 环境变量读取和运行参数解析 |
| `internal/app/server.go` | HTTP 路由、上传接口、订阅下载接口、缓存逻辑 |
| `internal/app/crypto.go` | 上传包加解密，必须和 workflow 的 OpenSSL 参数及文件格式对应 |
| `internal/app/archive.go` | `tar.gz` 解包和路径逃逸防护 |
| `internal/app/*_test.go` | 加解密兼容、上传解包、缓存下载和私密文件不暴露的测试 |
| `Dockerfile` | 多阶段构建，builder 使用 Go，运行时使用 `scratch` |
| `compose.yaml` | 本地容器运行示例，默认使用 named volume 挂载 `/data` |
| `.gitignore`、`.dockerignore` | 忽略本地密钥、数据目录、构建产物和压缩包 |

## HTTP 行为

| 接口 | 行为 |
|------|------|
| `POST /upload` | multipart 表单，字段 `token` 和文件字段 `file=config.tar.gz.aes` |
| `GET /sub` | `url` 参数必须以 `sub_secret` 内容开头；命中缓存时直接返回文件 |
| `GET /healthz` | 健康检查，返回 `ok` |

`/upload` 只接受文件名 `config.tar.gz.aes`。`/sub` 的 `target`、`ver` 和 secret 后缀只允许字母、数字、下划线和连字符，避免构造路径穿越。

## 数据目录

默认数据目录是当前目录，Docker 中是 `/data`。

| 文件或目录 | 说明 |
|------------|------|
| `upload_secret` | 单行 base64 AES 密钥，解码后必须是 32 字节 |
| `upload_token` | 上传 token |
| `sub_secret` | `/sub` 使用的 URL 前缀密钥 |
| `subscribe` | 一行一个订阅地址 |
| `config/` | 解包后的缓存配置 |

这些数据只由服务读取或写入，不作为静态文件暴露。

## 加解密约束

加解密必须和 `.github/workflows/subconverter.yml` 完全对应：

| 项 | 值 |
|----|----|
| 算法 | AES-256-CBC |
| padding | PKCS#7 |
| salt | 不使用，对应 `openssl enc -nosalt` |
| 上传包格式 | 前 16 字节为随机 IV，后续内容为密文 |
| key | `upload_secret` base64 解码后的 32 字节 |

改动 `crypto.go` 或 workflow 加密命令时，必须保留 OpenSSL 样例密文测试，并覆盖上传包 IV 前缀解析。

## 验证命令

```shell
go test ./...
go build ./cmd/config-depot
docker build -t config-depot:local-test .
docker compose config
```

如果本地沙箱不允许写默认 Go cache，可把 `GOCACHE` 指向仓库 `tmp/` 下的临时目录。端到端演练也放仓库 `tmp/` 下，生成的压缩包、密钥、上传结果和下载结果都应被 ignore 覆盖。
