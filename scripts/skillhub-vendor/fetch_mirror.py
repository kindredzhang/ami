#!/usr/bin/env python3
"""按 downloads.json 把文件下载到镜像目录，并核对校验和。

在一台能上外网的机器上跑，产出可以直接 rsync 到内网镜像服务器：

    python3 scripts/skillhub-vendor/fetch_mirror.py \
        skillhub-data/mirror-manifest/downloads.json \
        /tmp/dfmc-mirror

目录结构和 docs/skillhub.md 里的 base_url 一一对应：

    /tmp/dfmc-mirror/
        golang/dl/go1.27.3.linux-amd64.tar.gz
        rustup/dist/x86_64-unknown-linux-gnu/rustup-init
        anaconda/...
        tdf/...
        npm/...

下载完会重算 sha256，和清单比对；对不上直接报错，不留下坏文件。
"""
from __future__ import annotations

import hashlib
import json
import os
import sys
import urllib.request
from pathlib import Path

# 官方域名 -> 镜像目录前缀。和 docs/skillhub.md 的配置表对应。
ROUTES = [
    ("https://go.dev/dl/", "golang/dl/"),
    ("https://static.rust-lang.org/", "rustup/"),
    ("https://conda.anaconda.org/conda-forge/", "anaconda/conda-forge/"),
    ("https://api.anaconda.org/package/conda-forge/", "anaconda/package/conda-forge/"),
    ("https://github.com/", "github/"),
    ("https://objects.githubusercontent.com/", "github-objects/"),
    ("https://download.documentfoundation.org/", "tdf/"),
    ("https://registry.npmjs.org/", "npm/"),
]


def route_for(url: str) -> tuple[str, str] | None:
    """把官方 URL 映射成镜像内的相对路径。"""
    for prefix, target in ROUTES:
        if url.startswith(prefix):
            return target, url[len(prefix):]
    return None


def build_opener() -> urllib.request.OpenerDirector:
    proxies = {}
    for name in ("http_proxy", "https_proxy"):
        value = os.environ.get(name) or os.environ.get(name.upper())
        if value:
            proxies[name.split("_")[0]] = value
    return urllib.request.build_opener(urllib.request.ProxyHandler(proxies))


def download(opener, url: str, dest: Path, expect_sha: str) -> str:
    """下载并校验。返回 'ok' / 'skipped'。"""
    if dest.exists() and expect_sha:
        if sha256_of(dest) == expect_sha:
            return "skipped"

    dest.parent.mkdir(parents=True, exist_ok=True)
    tmp = dest.with_suffix(dest.suffix + ".part")
    try:
        with opener.open(url, timeout=180) as resp, open(tmp, "wb") as fh:
            while chunk := resp.read(1 << 20):
                fh.write(chunk)
    except Exception:
        tmp.unlink(missing_ok=True)
        raise

    if expect_sha:
        actual = sha256_of(tmp)
        if actual != expect_sha:
            tmp.unlink(missing_ok=True)
            raise SystemExit(f"校验和不匹配: {url}\n  期望 {expect_sha}\n  实际 {actual}")
    os.replace(tmp, dest)
    return "ok"


def sha256_of(path: Path) -> str:
    digest = hashlib.sha256()
    with open(path, "rb") as fh:
        while chunk := fh.read(1 << 20):
            digest.update(chunk)
    return digest.hexdigest()


def main() -> int:
    if len(sys.argv) < 3:
        print("usage: fetch_mirror.py <downloads.json> <mirror-dir>", file=sys.stderr)
        return 2

    manifest = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
    root = Path(sys.argv[2])
    opener = build_opener()

    ok = skipped = failed = 0
    for dep_id in sorted(manifest):
        entry = manifest[dep_id]
        for item in entry.get("files", []):
            url = item.get("url", "")
            if not url or url.startswith("apt://") or item.get("packages"):
                # apt 包和 npm/conda 的“整站镜像”条目没有单文件可下
                continue

            routed = route_for(url)
            if routed is None:
                print(f"  [未知来源] {dep_id}: {url}", file=sys.stderr)
                failed += 1
                continue
            target, rest = routed
            dest = root / target / rest

            try:
                state = download(opener, url, dest, item.get("sha256", ""))
                print(f"  [{state:7}] {dep_id:20} {dest.relative_to(root)}")
                ok += state == "ok"
                skipped += state == "skipped"
            except SystemExit:
                raise
            except Exception as exc:
                print(f"  [FAILED ] {dep_id:20} {url} — {exc}", file=sys.stderr)
                failed += 1

    print(f"\n完成: 下载 {ok}, 已有 {skipped}, 失败 {failed}")
    print(f"镜像目录: {root.resolve()}")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
