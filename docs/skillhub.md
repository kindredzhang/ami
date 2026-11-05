# 内部依赖源（skillhub）

内网部署的运行时/工具来源。替代公网的 `supermarket.memoh.ai` 依赖目录。

## 为什么要自建

云电脑里装的 rust、golang、node 这些依赖，**安装脚本本身是随依赖下发的**。
谁发布依赖，谁就决定在每台机器上执行什么。

用一个连不上（或不该连）的公网源，问题不只是装不上：

- 脚本内容由外部公司决定，你们没有审核记录
- 版本随上游漂移，环境不可复现
- 有问题的包无法追溯

自建源不是「网络通不通」的问题，是「谁有权往我的机器里写脚本」的问题。

## 两个源是分开的

```toml
[supermarket]
base_url = "https://supermarket.memoh.ai"        # 市场：skill、SaaS 应用
catalog_base_url = "http://skillhub:8090"        # 依赖目录：rust、go、node
# disabled = false                                # 连不上时设 true，列表返回空
```

`catalog_base_url` 留空则继承 `base_url`，单源部署行为不变。

## 发布一个依赖

```
skillhub-data/
└── dependencies/
    └── <id>/
        ├── dependency.yaml   清单（权威）
        ├── artifact.tar.gz   安装包，必须包含同一份清单
        └── icon.svg          可选图标，只接受 SVG
```

放好之后重启容器：

```bash
docker compose restart skillhub
```

服务启动时会校验全部依赖，任何一个不合法就**拒绝启动**并指出是哪个目录。
这比启动后吐出一份客户端会拒绝的目录要好——错误在发布时暴露，不在用户装机时暴露。

### 打归档包

`artifact.tar.gz` 不进版本库（它是构建产物），改完清单或脚本后跑：

```bash
./skillhub-data/build.sh
```

脚本会按每个依赖的实际文件重新打包。归档里必须是**扁平文件名**，
不能有子目录。

macOS 上打包务必带 `COPYFILE_DISABLE=1`，否则 tar 会把 AppleDouble 的
`._xxx` 元数据文件一起打进去——那些文件不在清单里，校验会直接拒掉整个依赖。
这个坑在 Mac 上必现，在 Linux 上不出现，`build.sh` 已经处理了。

### 清单示例

```yaml
schema_version: "1"
id: hello-tool
name: Hello Tool
category: tool          # agent | runtime | tool
icon: icon.svg
description: 一句话说明。
source: managed         # 或 image（表示镜像里已有基线）

provides:
  - hello-tool

platforms:
  - { os: linux, arch: [amd64, arm64], libc: glibc }

timeouts:
  install: 120
  remove: 60
  check_update: 30
  version: 30

scripts:
  install: install.sh
  remove: remove.sh
  check_update: check-update.sh
```

`scripts` 里点到的文件必须在归档中存在，且归档里不能出现清单未声明的文件。

## 三个容易踩的坑

这三个都会让发布出来的目录被客户端拒绝，而且报错不直观：

1. **`manifest_digest` 不是清单的哈希。** 它把清单和 `scripts` 点到的每个脚本文件
   一起算，不含图标。服务端会自动算，不用手写。
2. **`archive_size` 是解压后的 tar 长度**（含 512 字节头部和对齐填充），
   不是 `artifact.tar.gz` 的大小；`uncompressed_size` 才是文件内容之和。
3. **revision 是 release 文档的 sha256**，不是安装包的哈希。
   客户端会用它校验 metadata，服务端从实际发出的字节推导。

## 校验

不启动服务，只检查数据目录：

```bash
docker compose run --rm skillhub -check -data /data
# 或本机
go run ./cmd/skillhub -check -data ./skillhub-data
```

输出每个依赖的 id、版本、revision，最后给出整个快照的 revision。

## 接口

客户端只调用这四个：

| 路径 | 说明 |
|---|---|
| `GET /api/dependencies?registry=ami&page=N&limit=M` | 分页索引 |
| `GET /api/registries/ami/dependencies/{id}/releases/{revision}` | release 文档 |
| `GET /api/artifacts/dependency/{digest}` | 安装包 |
| `GET /api/artifacts/icon/{digest}` | 图标 |

外加 `GET /healthz` 供容器健康检查。

协议细节以 `internal/workspacedeps/catalog/release.go` 为准——本服务直接调它的
解码函数做自检，所以两边不会漂移。

## 已验证

用 ami 自己的三个解码器（`DecodeIndex`、`DecodeRelease`、`FromArtifact`）
走了一遍完整链路：拉索引 → 拉 release → 拉安装包 → 校验清单与归档一致性 → 取图标。
全部通过。这一步是必要的：协议里有若干处「看起来对但不合法」的写法，
只有真解码才算数。
