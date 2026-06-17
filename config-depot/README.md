# config-depot

`config-depot` 是 SubConfig 的发布服务。它接收 GitHub Actions 上传的加密配置包，解密解包后，通过 `/sub` 给客户端提供缓存配置下载。

## 数据目录

默认数据目录是 `data`，Docker 镜像中是 `/data`。

| 文件或目录 | 说明 |
|------------|------|
| `upload_secret` | 单行 base64 AES 密钥，解码后必须是 32 字节 |
| `upload_token` | `POST /upload` 使用的上传 token |
| `sub_secret` | `GET /sub` 使用的 URL 前缀密钥 |
| `subscribe` | 一行一个订阅地址 |
| `config/` | 解包后的配置缓存目录 |

## 运行

本地运行：

```shell
go run ./cmd/config-depot
```

Docker Compose 运行：

```shell
docker compose up -d --build
```

compose 默认把本地 `./data` 绑定挂载到容器的 `/data`。容器以非 root 用户运行；正式部署时如果改成绑定原来的站点数据目录，例如 `/var/www/sub:/data`，需要先让容器用户可写该目录。

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `CONFIG_DEPOT_ADDR` | `:8080` | HTTP 监听地址 |
| `CONFIG_DEPOT_DATA_DIR` | `data` | 数据目录 |
| `CONFIG_DEPOT_CONFIG_DIR` | `<data>/config` | 配置缓存目录 |
| `CONFIG_DEPOT_UPLOAD_SECRET_FILE` | `<data>/upload_secret` | 上传密钥文件 |
| `CONFIG_DEPOT_UPLOAD_TOKEN_FILE` | `<data>/upload_token` | 上传 token 文件 |
| `CONFIG_DEPOT_SUB_SECRET_FILE` | `<data>/sub_secret` | 订阅密钥文件 |
| `CONFIG_DEPOT_SUBSCRIBE_FILE` | `<data>/subscribe` | 订阅列表文件 |
| `CONFIG_DEPOT_BACKEND_URL` | `https://subc.020.name/sub?` | 缓存未命中时使用的 subconverter 后端 |
| `CONFIG_DEPOT_REMOTE_CONFIG_URL` | `https://github.com/AoEiuV020/SubConfig/raw/main/subconverter.ini` | 默认远程配置地址 |

## 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/upload` | `POST` | multipart 表单，字段 `token` 和文件字段 `file=config.tar.gz.aes` |
| `/sub` | `GET` | 参数 `url=<sub_secret>`，可选 `target`、`ver`、`cache=false` |
| `/healthz` | `GET` | 返回 `ok` |

`/upload` 解密方式与 workflow 完全对应：上传包前 16 字节为随机 IV，后续内容为 AES-256-CBC 密文；padding 为 PKCS#7，`openssl enc` 使用 `-nosalt`。
