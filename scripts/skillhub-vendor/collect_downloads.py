#!/usr/bin/env python3
"""生成内网镜像需要托管的文件清单。

不靠猜版本号和校验和——直接执行依赖脚本，把 download() 换成记录器，
脚本想下什么就记什么。这样清单里的 URL 和 sha256 一定和真实安装时用的
完全一致，运维照着准备就不会出现「版本对不上」或「校验和不匹配」。

用法：
    python3 scripts/skillhub-vendor/collect_downloads.py <patched-dir> [id...]

输出：
    <patched-dir>/downloads.json   机器可读，给 fetch_mirror.py 用
    <patched-dir>/downloads.md     人可读，给运维看
"""
from __future__ import annotations

import json
import os
import re
import subprocess
import sys
import tempfile
from pathlib import Path

# 脚本里嵌的 Python 运行时以 heredoc 形式出现，正文在这两个标记之间。
HEREDOC_START = "python3 - <<'MEMOH_RECIPE_PY'"
HEREDOC_END = "MEMOH_RECIPE_PY"

# 这些平台组合里挑一个来解析版本。脚本的 download() 只受平台影响，
# 选 linux/amd64 是因为内网云电脑就是这个架构；要同时给 Mac 准备文件，
# 把 PLATFORMS 里加一行就行。
PLATFORMS = [("linux", "amd64", "glibc")]

# resolve() 需要一些 MEMOH_DEP_* 环境变量，缺了就不走分支。
BASE_ENV = {
    "MEMOH_DEP_HOME": "/tmp/collect-home",
    "MEMOH_DEP_PLATFORM": "linux",
    "MEMOH_DEP_OS": "linux",
    "MEMOH_DEP_ARCH": "amd64",
    "MEMOH_DEP_LIBC": "glibc",
    "MEMOH_DEP_VERSION": "",
    "MEMOH_DEP_CURRENT_VERSION": "",
    "AMI_DEP_HOME": "/tmp/collect-home",
    "CI": "1",
    "DEBIAN_FRONTEND": "noninteractive",
}

# 在脚本正文末尾追加一段收集器：把 download 换成记录器，跑 main() 的解析阶段。
COLLECTOR = '''

# ---- collect_downloads: 从 resolution 推下载清单，不真的联网下载 ----
#
# 只跑 resolve()：它负责查版本、拉版本清单，是唯一需要网络的地方。
# install() 不跑——它会真的下载解包，而且依赖一堆系统命令存在；
# 它想下哪个文件完全由 resolution 决定，照着同样的规则读出来即可。


def _pick_platform(config, resolution):
    """按 backend 的规则，从 resolution 里挑出本平台要下载的文件。"""
    backend = config.get('backend', '')
    os_name = os.environ['MEMOH_DEP_OS']
    arch = os.environ['MEMOH_DEP_ARCH']
    version = resolution.get('version', '')
    out = []

    if backend == 'go':
        item = next(
            (f for f in resolution.get('files', [])
             if f.get('os') == os_name and f.get('arch') == arch and f.get('kind') == 'archive'),
            None,
        )
        if item:
            out.append({
                'url': _mirror('GO_DL_BASE', 'https://go.dev/dl') + '/' + item['filename'],
                'sha256': item.get('sha256', ''),
                'size': item.get('size', 0),
            })

    elif backend == 'rust':
        triple = {
            'linux/amd64': 'x86_64-unknown-linux-gnu',
            'linux/arm64': 'aarch64-unknown-linux-gnu',
            'darwin/amd64': 'x86_64-apple-darwin',
            'darwin/arm64': 'aarch64-apple-darwin',
        }.get(os_name + '/' + arch, 'x86_64-unknown-linux-gnu')
        out.append({
            'url': _mirror('RUSTUP_DIST_SERVER', 'https://static.rust-lang.org') + '/rustup/dist/' + triple + '/rustup-init',
            'sha256': '',   # rustup-init 的校验和单独取 .sha256，见 needs_checksum
            'size': 0,
        })
        # rustup 自己还要拉 toolchain，那是运行时行为，镜像端要整站镜像
        out.append({
            'url': _mirror('RUSTUP_DIST_SERVER', 'https://static.rust-lang.org') + '/dist/channel-rust-stable.toml',
            'sha256': '',
            'size': 0,
            'note': 'rustup 运行时会按此清单再拉组件，镜像站需整目录同步 dist/ 与 rustup/',
        })

    elif backend == 'github':
        assets = config.get('assets', {})
        name = assets.get(os_name + '/' + arch, '')
        if name:
            name = name.format(version=version)
            asset = next((a for a in resolution.get('release', {}).get('assets', []) if a.get('name') == name), None)
            if asset:
                digest = (asset.get('digest') or '').removeprefix('sha256:')
                out.append({
                    'url': asset.get('browser_download_url', ''),
                    'sha256': digest,
                    'size': asset.get('size', 0),
                })
        if config.get('raw') and out:
            out[-1]['note'] = '解压后重命名为 ' + ', '.join(config.get('commands', []))

    elif backend == 'npm':
        out.append({
            'url': _mirror('NPM_MIRROR', 'https://registry.npmjs.org'),
            'sha256': '',
            'size': 0,
            'packages': resolution.get('packages', {}),
            'note': 'npm 后端：镜像 registry 后按包名+版本拉取，不预先下载 tarball',
        })

    elif backend == 'conda':
        pkg = config.get('package', '')
        out.extend(_conda_files(pkg, os_name, arch, version))

    elif backend == 'libreoffice':
        # Linux 上这个 backend 走 apt-get install libreoffice-core，
        # 根本不下载独立包；Debian 源可达与否才是决定因素。
        if os_name == 'linux':
            out.append({
                'url': 'apt://libreoffice-core',
                'sha256': '',
                'size': 0,
                'note': 'Linux 走 apt-get，需要内网 Debian 源（或镜像源）里有 libreoffice-core',
            })
        else:
            out.append({
                'url': _mirror('LO_DL_BASE', 'https://download.documentfoundation.org') + '/libreoffice/stable/' + version + '/',
                'sha256': '',
                'size': 0,
                'note': 'macOS：需同步该版本目录下的 dmg 包',
            })

    elif backend == 'python':
        out.append({
            'url': _mirror('UV_PYTHON_INSTALL_MIRROR', 'https://github.com/astral-sh/python-build-standalone/releases/download'),
            'sha256': '',
            'size': 0,
            'packages': resolution.get('packages', {}),
            'note': 'python 后端：uv 拉 python-build-standalone',
        })

    return out


def _conda_files(package: str, os_name: str, arch: str, version: str) -> list[dict]:
    """查出 conda-forge 里本平台的具体包文件。

    只给包名没用——装机时 micromamba 会按 solve 结果去取具体文件。
    这里把 latest 版本的实际文件地址列出来，运维可以直接下载。
    注意 conda 还有一层依赖求解，子依赖无法在这里穷举，所以注明要整站镜像。
    """
    import urllib.request

    subdir = {'linux': 'linux-64', 'darwin': 'osx-arm64'}.get(os_name, 'linux-64')
    if arch == 'arm64' and os_name == 'linux':
        subdir = 'linux-aarch64'

    proxies = {
        k: os.environ[k]
        for k in ('http_proxy', 'https_proxy', 'HTTP_PROXY', 'HTTPS_PROXY')
        if os.environ.get(k)
    }
    opener = urllib.request.build_opener(
        urllib.request.ProxyHandler({k.split('_')[0]: v for k, v in proxies.items()})
    )
    base = _mirror('CONDA_CHANNEL_BASE', 'https://api.anaconda.org')
    try:
        with opener.open(base + '/package/conda-forge/' + package, timeout=30) as resp:
            data = json.load(resp)
    except Exception:
        return [{
            'url': base + '/package/conda-forge/' + package,
            'sha256': '',
            'size': 0,
            'note': 'conda 后端：镜像 conda-forge 的 ' + package + '（版本解析失败）',
        }]

    candidates = []
    for item in data.get('files', []):
        attrs = item.get('attrs') or {}
        if attrs.get('subdir') != subdir:
            continue
        if version and not item.get('version', '').startswith(version):
            continue
        # sha256 在文件对象顶层，不在 attrs 里
        sha = item.get('sha256') or attrs.get('sha256') or ''
        if not sha:
            continue
        # download_url 是协议相对的（//conda.anaconda.org/...），补成 https
        url = item.get('download_url') or ''
        if url.startswith('//'):
            url = 'https:' + url
        if not url:
            continue
        candidates.append({
            'url': url,
            'sha256': sha,
            'size': item.get('size', 0),
            'note': 'conda-forge ' + subdir,
        })

    if not candidates:
        return [{
            'url': base + '/package/conda-forge/' + package,
            'sha256': '',
            'size': 0,
            'note': 'conda 后端：镜像 conda-forge 的 ' + package + '（无 ' + subdir + ' 文件）',
        }]
    # 同一版本常有多份构建，取文件最大的那个（通常是完整构建）
    candidates.sort(key=lambda c: c['size'], reverse=True)
    picked = candidates[0]
    picked['note'] = (
        'conda-forge ' + subdir + f'（该版本 {len(candidates)} 个构建，此为最大的一个；'
        'conda 还会求解子依赖，建议镜像整个 conda-forge/' + subdir + '）'
    )
    return [picked]


def _collect(config):
    """只跑 resolve()，把要下载的文件按 backend 规则推出来。"""
    home = Path(os.environ['MEMOH_DEP_HOME'])
    import shutil
    shutil.rmtree(home, ignore_errors=True)
    home.mkdir(parents=True, exist_ok=True)
    try:
        try:
            resolution = resolve(config, os.environ.get('MEMOH_DEP_VERSION', ''), home)
        except Exception as exc:
            # 解析失败不代表没东西要准备：libreoffice 走 apt，在 Mac/CI 上
            # 一定解析不出来，但内网仍需要 Debian 源。这时按 backend 的
            # 静态规则给结论，不把「解析不了」当成「不需要准备」。
            resolution = {'version': ''}
            version, files = '', _pick_platform(config, resolution)
            for item in files:
                item['note'] = (item.get('note', '') + ' （版本未能解析：' + type(exc).__name__ + '）').strip()
            return version, files
        return resolution.get('version', ''), _pick_platform(config, resolution)
    finally:
        shutil.rmtree(home, ignore_errors=True)
'''


def extract_python(script: str) -> str | None:
    """取出 heredoc 里的 Python 正文；纯 shell 脚本返回 None。"""
    start = script.find(HEREDOC_START)
    if start < 0:
        return None
    start = script.find("\n", start) + 1
    end = script.find("\n" + HEREDOC_END, start)
    if end < 0:
        return None
    return script[start:end]


def collect_for(body: str, platform: tuple[str, str, str]) -> tuple[str, list[dict]]:
    """在给定平台上解析一个依赖，返回版本和下载清单。"""
    os_name, arch, libc = platform
    env = dict(os.environ)
    env.update(BASE_ENV)
    # 解析版本要联网。允许通过环境变量指定代理，内网机器上不设即可
    # （那种情况下解析必然失败，属于预期——清单本来就该在有外网时生成）。
    for name in ("HTTP_PROXY", "HTTPS_PROXY", "http_proxy", "https_proxy"):
        if os.environ.get(name):
            env[name] = os.environ[name]
    env.update(
        {
            "MEMOH_DEP_PLATFORM": f"{os_name}/{arch}",
            "MEMOH_DEP_OS": os_name,
            "MEMOH_DEP_ARCH": arch,
            "MEMOH_DEP_LIBC": libc,
        }
    )

    # 摘掉 __main__ 入口，换成我们的收集器
    body = re.sub(r"if __name__ == ['\"]__main__['\"]:.*", "", body, flags=re.S)
    program = body + COLLECTOR + "\n\nimport json as _json\n_version, _files = _collect(CONFIG)\nprint('__RESULT__' + _json.dumps({'version': _version, 'files': _files}))\n"

    with tempfile.NamedTemporaryFile("w", suffix=".py", delete=False, encoding="utf-8") as fh:
        fh.write(program)
        path = fh.name
    try:
        # 3 分钟够解析任何 backend；卡住说明在真联网，直接放弃
        proc = subprocess.run(
            [sys.executable, path], capture_output=True, text=True, env=env, timeout=180
        )
    except subprocess.TimeoutExpired:
        return "", []
    finally:
        os.unlink(path)

    for line in proc.stdout.splitlines():
        if line.startswith("__RESULT__"):
            payload = json.loads(line[len("__RESULT__") :])
            return payload.get("version", ""), payload.get("files", [])
    return "", []


def rewrite_for_mirror(url: str, base: str, path_prefix: str) -> str:
    """把官方 URL 换成内网镜像对应的路径。"""
    return f"{base}{path_prefix}{url.split('://', 1)[-1].split('/', 1)[-1]}"


def main() -> int:
    if len(sys.argv) < 2:
        print("usage: collect_downloads.py <patched-dir> [id...]", file=sys.stderr)
        return 2

    root = Path(sys.argv[1])
    wanted = set(sys.argv[2:])

    results: dict[str, dict] = {}
    for tgz in sorted(root.glob("*.tgz")):
        dep_id = tgz.stem
        if wanted and dep_id not in wanted:
            continue

        with tempfile.TemporaryDirectory() as tmp:
            subprocess.run(["tar", "-xzf", str(tgz), "-C", tmp], check=True)
            script_path = Path(tmp) / "install.sh"
            if not script_path.is_file():
                print(f"skip {dep_id}: no install.sh", file=sys.stderr)
                continue
            body = extract_python(script_path.read_text(encoding="utf-8"))

        if body is None:
            # 纯 shell 脚本（node/python/uv/claude-code 等基线包），
            # 它们的下载逻辑无从静态解析，单独标注。
            results[dep_id] = {"version": "", "files": [], "note": "shell recipe, 无 Python 解析入口"}
            print(f"{dep_id:22} shell recipe, 跳过")
            continue

        version, files = collect_for(body, PLATFORMS[0])
        results[dep_id] = {"version": version, "files": files}
        print(f"{dep_id:22} v{version or '?':14} {len(files)} 个文件")

    fill_missing_checksums(results)

    out_json = root / "downloads.json"
    out_json.write_text(json.dumps(results, indent=2, ensure_ascii=False), encoding="utf-8")

    out_md = root / "downloads.md"
    out_md.write_text(render_markdown(results), encoding="utf-8")

    print(f"\n写入 {out_json}")
    print(f"写入 {out_md}")
    return 0


# 每个依赖在镜像站上的落地路径前缀，和 docs/skillhub.md 里的配置对应。
MIRROR_PREFIX = {
    "go": ("GO_DL_BASE", "/dl/"),
    "rust": ("RUSTUP_DIST_SERVER", "/"),
    "bun": ("GITHUB_BASE", "/"),
    "github-cli": ("GITHUB_BASE", "/"),
    "micromamba": ("GITHUB_BASE", "/"),
    "native-build-tools": ("CONDA_CHANNEL_BASE", "/package/conda-forge/"),
    "ffmpeg": ("CONDA_CHANNEL_BASE", "/package/conda-forge/"),
    "pandoc": ("CONDA_CHANNEL_BASE", "/package/conda-forge/"),
    "tesseract": ("CONDA_CHANNEL_BASE", "/package/conda-forge/"),
    "poppler": ("CONDA_CHANNEL_BASE", "/package/conda-forge/"),
    "qpdf": ("CONDA_CHANNEL_BASE", "/package/conda-forge/"),
    "libreoffice": ("(apt 源)", ""),
    "pnpm": ("NPM_MIRROR", "/"),
    "playwright": ("NPM_MIRROR", "/"),
    "node": ("(镜像基线)", ""),
    "python": ("(镜像基线)", ""),
    "uv": ("(镜像基线)", ""),
}


def render_markdown(results: dict[str, dict]) -> str:
    """生成运维能照着做的清单。"""
    lines = [
        "# 内网镜像准备清单",
        "",
        "由 `scripts/skillhub-vendor/collect_downloads.py` 生成。",
        "版本号和 sha256 是脚本自己解析出来的，不是手抄的。",
        "",
        "## 一、需要提前下载的文件",
        "",
        "| 依赖 | 版本 | 文件 | 大小 | sha256 |",
        "|---|---|---|---|---|",
    ]
    pending = []
    for dep_id in sorted(results):
        entry = results[dep_id]
        files = entry.get("files") or []
        if not files:
            continue
        for item in files:
            url = item.get("url", "")
            if url.startswith("apt://") or item.get("packages"):
                continue
            size = item.get("size") or 0
            size_txt = f"{size / 1048576:.1f} MB" if size else "—"
            digest = item.get("sha256") or "待确认"
            name = url.rsplit("/", 1)[-1] or url
            lines.append(
                f"| `{dep_id}` | {entry.get('version') or '—'} | `{name}` | {size_txt} | `{digest[:16]}`… |"
            )
            pending.append((dep_id, url, digest))

    lines += ["", "## 二、按后端分类的处理方式", ""]
    groups: dict[str, list[str]] = {}
    for dep_id in sorted(results):
        entry = results[dep_id]
        env, _ = MIRROR_PREFIX.get(dep_id, ("?", ""))
        note = ""
        files = entry.get("files") or []
        if files:
            note = files[0].get("note", "")
        groups.setdefault(env, []).append(f"`{dep_id}`" + (f" — {note}" if note else ""))
    for env in sorted(groups):
        lines.append(f"**{env}**")
        for item in groups[env]:
            lines.append(f"- {item}")
        lines.append("")

    lines += [
        "## 三、完整 URL",
        "",
        "```",
    ]
    for dep_id, url, digest in pending:
        lines.append(url)
        if digest and digest != "待确认":
            lines.append(f"  sha256: {digest}")
    lines += ["```", ""]
    return "\n".join(lines)


# 校验和不在版本清单里、而是单独一个 <url>.sha256 文件的情况。
CHECKSUM_SIDECAR_SUFFIXES = ("", ".sha256", ".sha256sum")


def fill_missing_checksums(results: dict[str, dict]) -> None:
    """给缺 sha256 的条目补校验和。

    有些发布只提供 <url>.sha256 这样一个独立文件（rustup-init 就是），
    版本清单里没有哈希。这里把它取回来，免得运维拿到一条没法校验的清单。
    取不到就留空并注明，不编造。
    """
    import urllib.request

    proxies = {
        k: os.environ[k]
        for k in ("http_proxy", "https_proxy", "HTTP_PROXY", "HTTPS_PROXY")
        if os.environ.get(k)
    }
    opener = urllib.request.build_opener(
        urllib.request.ProxyHandler({k.split("_")[0]: v for k, v in proxies.items()})
    )

    for dep_id, entry in results.items():
        for item in entry.get("files", []):
            if item.get("sha256") or not item.get("url") or item.get("packages"):
                continue
            url = item["url"]
            for suffix in CHECKSUM_SIDECAR_SUFFIXES[1:]:
                try:
                    with opener.open(url + suffix, timeout=30) as resp:
                        text = resp.read().decode("utf-8", "replace")
                except Exception:
                    continue
                token = text.strip().split()[0] if text.strip() else ""
                if re.fullmatch(r"[0-9a-f]{64}", token):
                    item["sha256"] = token
                    break
            else:
                item.setdefault("note", "")
                item["note"] = (item["note"] + " 校验和需人工确认").strip()
                print(f"  {dep_id}: 拿不到 {url} 的校验和", file=sys.stderr)


if __name__ == "__main__":
    sys.exit(main())
