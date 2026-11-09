# 内网镜像准备清单

由 `scripts/skillhub-vendor/collect_downloads.py` 生成。
版本号和 sha256 是脚本自己解析出来的，不是手抄的。

## 一、需要提前下载的文件

| 依赖 | 版本 | 文件 | 大小 | sha256 |
|---|---|---|---|---|
| `bun` | 1.4.4 | `bun-linux-x64-baseline.zip` | 49.4 MB | `e0446eb55995772e`… |
| `ffmpeg` | 9.0.2 | `ffmpeg-9.0.2-gpl_h8b97b3f_502.conda` | 13.2 MB | `9b60f40e151afed3`… |
| `github-cli` | 2.103.3 | `gh_2.103.3_linux_amd64.tar.gz` | 14.3 MB | `4c806978ffce1cec`… |
| `go` | 1.27.3 | `go1.27.3.linux-amd64.tar.gz` | 67.3 MB | `65efbc80fdc1fa2c`… |
| `micromamba` | 2.10.0-0 | `micromamba-linux-64` | 18.1 MB | `701f39c06f8e6a67`… |
| `native-build-tools` | 2.0.0 | `c-compiler-2.0.0-ha29393d_0.conda` | 0.0 MB | `a54488d2b00380f5`… |
| `pandoc` | 3.13 | `pandoc-3.13-ha770c72_0.conda` | 22.4 MB | `1aeedb498ad58c8f`… |
| `poppler` | 26.10.0 | `poppler-26.10.0-h7813f90_0.conda` | 2.0 MB | `a36b80002c202621`… |
| `qpdf` | 12.4.2 | `qpdf-12.4.2-h7db8e16_1.conda` | 1.5 MB | `172d9f44efb8dc42`… |
| `rust` | 1.99.1 | `rustup-init` | — | `b1c8cb14e30889e7`… |
| `rust` | 1.99.1 | `channel-rust-stable.toml` | — | `待确认`… |
| `tesseract` | — | `tesseract-5.5.3-h7618cdf_0.conda` | 172.6 MB | `62ccb114986026a7`… |

## 二、按后端分类的处理方式

**(apt 源)**
- `libreoffice` — Linux 走 apt-get，需要内网 Debian 源（或镜像源）里有 libreoffice-core （版本未能解析：FileNotFoundError） 校验和需人工确认

**(镜像基线)**
- `node`
- `python`
- `uv`

**CONDA_CHANNEL_BASE**
- `ffmpeg` — conda-forge linux-64（该版本 48 个构建，此为最大的一个；conda 还会求解子依赖，建议镜像整个 conda-forge/linux-64）
- `native-build-tools` — conda-forge linux-64（该版本 1 个构建，此为最大的一个；conda 还会求解子依赖，建议镜像整个 conda-forge/linux-64）
- `pandoc` — conda-forge linux-64（该版本 1 个构建，此为最大的一个；conda 还会求解子依赖，建议镜像整个 conda-forge/linux-64）
- `poppler` — conda-forge linux-64（该版本 1 个构建，此为最大的一个；conda 还会求解子依赖，建议镜像整个 conda-forge/linux-64）
- `qpdf` — conda-forge linux-64（该版本 2 个构建，此为最大的一个；conda 还会求解子依赖，建议镜像整个 conda-forge/linux-64）
- `tesseract` — conda-forge linux-64（该版本 17 个构建，此为最大的一个；conda 还会求解子依赖，建议镜像整个 conda-forge/linux-64） （版本未能解析：TimeoutError）

**GITHUB_BASE**
- `bun`
- `github-cli`
- `micromamba` — 解压后重命名为 micromamba

**GO_DL_BASE**
- `go`

**NPM_MIRROR**
- `playwright` — npm 后端：镜像 registry 后按包名+版本拉取，不预先下载 tarball
- `pnpm` — npm 后端：镜像 registry 后按包名+版本拉取，不预先下载 tarball

**RUSTUP_DIST_SERVER**
- `rust`

## 三、完整 URL

```
https://github.com/oven-sh/bun/releases/download/bun-v1.4.4/bun-linux-x64-baseline.zip
  sha256: e0446eb55995772eb9894bebae9a6be368d985a8b9271ee124486f9946eedd76
https://conda.anaconda.org/conda-forge/linux-64/ffmpeg-9.0.2-gpl_h8b97b3f_502.conda
  sha256: 9b60f40e151afed3b7d46778982675ec0195c4d58747e978b7d62260ad928ef9
https://github.com/cli/cli/releases/download/v2.103.3/gh_2.103.3_linux_amd64.tar.gz
  sha256: 4c806978ffce1cec70853302a56d70ac6da4f19bae7adca4e1a8ab00633dd91d
https://go.dev/dl/go1.27.3.linux-amd64.tar.gz
  sha256: 65efbc80fdc1fa2cd2631cad5b481777a68e8725f34e89f50fc6ac714e379b40
https://github.com/mamba-org/micromamba-releases/releases/download/2.10.0-0/micromamba-linux-64
  sha256: 701f39c06f8e6a670a0e186c8479b9e5c0502a2f0cda6cef65d83bdd2875e8bd
https://conda.anaconda.org/conda-forge/linux-64/c-compiler-2.0.0-ha29393d_0.conda
  sha256: a54488d2b00380f542a2e1aa00dee3d878803a8c2c025dfce852a23685584dc6
https://conda.anaconda.org/conda-forge/linux-64/pandoc-3.13-ha770c72_0.conda
  sha256: 1aeedb498ad58c8f78176d5aaad11e32501d78dd6cdcd842e1a6d5c308e1830c
https://conda.anaconda.org/conda-forge/linux-64/poppler-26.10.0-h7813f90_0.conda
  sha256: a36b80002c202621dd0e1cd7d6f388d11b7013b2dda7ce1f31d934669535f2c5
https://conda.anaconda.org/conda-forge/linux-64/qpdf-12.4.2-h7db8e16_1.conda
  sha256: 172d9f44efb8dc42b8f8dc8725f2c1cbff976b28826b9a494446bdb38d3aa53d
https://static.rust-lang.org/rustup/dist/x86_64-unknown-linux-gnu/rustup-init
  sha256: b1c8cb14e30889e7e2999e6b69e62d9e9b9ffc0f4df8180564d306d9680e6c55
https://static.rust-lang.org/dist/channel-rust-stable.toml
https://conda.anaconda.org/conda-forge/linux-64/tesseract-5.5.3-h7618cdf_0.conda
  sha256: 62ccb114986026a75329b3060f1a77327a699644c8fb5aed322c14fd949b5273
```
