# SubConfig

SubConfig 是一组 subconverter 外部配置，并通过 GitHub Actions 定时生成订阅转换结果。

## 使用方式

fork 后可以直接运行 workflow 试用。未配置 `SUBSCRIBE` 时，workflow 会用示例节点生成配置，并上传到 Actions artifact 供检查。

需要发布到自己的服务时，配置以下 secrets：

| 名称 | 说明 |
|------|------|
| `SUBSCRIBE` | 订阅链接，一行一个 |
| `UPLOAD_SECRET` | 配置包 AES-256-CBC 加密密钥，base64 字符串，解码后 32 字节；上传包前 16 字节保存随机 IV |
| `UPLOAD_TOKEN` | 上传接口使用的 token |
| `DEPLOY_URL` | `config-depot` 上传地址，例如 `https://host/upload` |

生成密钥示例：

```shell
head -c 32 /dev/urandom | base64 > upload_secret
head -c 32 /dev/urandom | od -A n -v -t x1 | tr -d ' \n' > upload_token
```

四个 secrets 都存在时，workflow 才会发布到 `DEPLOY_URL`。

## 下载配置

`config-depot` 提供下载入口：

```text
https://host/sub?url=<secret>
https://host/sub?url=<secret>-basic
https://host/sub?target=quan&url=<secret>-basic
```

`secret` 是服务器上的 `sub_secret` 内容。后面追加 `-basic`、`-noban`、`-ban` 等后缀时，会读取对应的配置变体和缓存文件。

## 部署服务

发布服务在 `config-depot/`：

[config-depot/README.md](config-depot/README.md)

Nginx 可以整体反代到 Go 服务：

```nginx
location / {
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_pass http://127.0.0.1:8080;
}
```
