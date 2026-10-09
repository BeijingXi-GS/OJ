# 本地开发 MinIO 镜像

仅用于 `linux/amd64` 开发 engine。MinIO 源码提交与证书来源摘要固定在 [source-lock.json](source-lock.json)，方案与限制见 [ADR-0002](../../docs/adr/0002-local-minio-source-image.md)。

在仓库根目录运行：

```powershell
npm run infra:build-minio
npm run infra:config
npm run infra:up
node scripts/infra-smoke.mjs
```

前置：本机 Go 1.25+、Docker Linux amd64 engine、模块源 / checksum 服务可访问。构建器使用项目 `.local/` 缓存，并保留 Go 校验数据库。没有使用 `GOSUMDB=off`。默认模块代理顺序为 goproxy.cn → proxy.golang.org → direct，可通过当次 `GOPROXY` 覆盖。

首次需要取得固定证书来源镜像；构建脚本会在没有缓存时拉取，也可预先执行：

```powershell
docker pull redis:7.4.5-alpine@sha256:bb186d083732f669da90be8b0f975a37812b15e913465bb14d845db72a4e3e08
```

Dockerfile 从该镜像只复制 CA 文件，最终运行镜像为 scratch，含 MinIO、健康探针、CA 和上游许可证，以 UID/GID 10001 运行。编译的是上游固定提交，不是原 Compose 中的 2025-09-07 二进制版本。

`.local/minio-build/build-metadata.json` 保存来源版本、module 校验值、Go 版本和二进制 SHA-256；不要把构建产物、缓存或 `.env` 提交到仓库。普通客户端仍使用 localhost 的 S3 9000 / Console 9001，不向前端暴露 root 凭据。

源码仓库已归档，此构建不承诺后续维护，也不替代生产存储 / 安全审阅。冷机器初始化仍依赖模块下载与固定 CA 镜像，不声称完全离线。
