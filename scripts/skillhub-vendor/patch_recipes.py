#!/usr/bin/env python3
"""把上游依赖脚本里的硬编码下载地址改成可配置的镜像前缀。

上游脚本把 go.dev / static.rust-lang.org / api.anaconda.org 这些域名写死在
字符串拼接里，没有任何环境变量入口。内网没有外网访问，原样搬到 skillhub 也没用。

这里的做法是：在每个 URL 常量的位置插入一个 os.environ.get 兜底，默认值仍是
上游原地址。不配环境变量时行为完全不变，配了才走内网镜像。
"""
import os
import re
import shutil
import subprocess
import sys
import tempfile

# (正则, 替换模板, 需要注入的环境变量名)
# 模板里的 {env} 会被换成读取环境变量的表达式。
PATCHES = [
    # (匹配模式, 替换内容, 依赖的环境变量名, 默认地址)
    # 匹配模式要把原有引号一起吃掉，替换内容自己带引号，否则会嵌进字符串里。
    (
        r"'https://go\.dev/dl/'",
        "_mirror('GO_DL_BASE', 'https://go.dev/dl') + '/'",
        "GO_DL_BASE",
    ),
    (
        r"'https://go\.dev/dl/\?mode=json'",
        "_mirror('GO_DL_BASE', 'https://go.dev/dl') + '/?mode=json'",
        "GO_DL_BASE",
    ),
    (
        r"f'https://static\.rust-lang\.org/dist/channel-rust-\{channel\}\.toml'",
        "f\"{_mirror('RUSTUP_DIST_SERVER', 'https://static.rust-lang.org')}/dist/channel-rust-{channel}.toml\"",
        "RUSTUP_DIST_SERVER",
    ),
    (
        r"'https://static\.rust-lang\.org/rustup/dist/'",
        "_mirror('RUSTUP_DIST_SERVER', 'https://static.rust-lang.org') + '/rustup/dist/'",
        "RUSTUP_DIST_SERVER",
    ),
    (
        r"'https://api\.anaconda\.org/package/conda-forge/'",
        "_mirror('CONDA_CHANNEL_BASE', 'https://api.anaconda.org') + '/package/conda-forge/'",
        "CONDA_CHANNEL_BASE",
    ),
    (
        r"'https://download\.documentfoundation\.org/libreoffice/stable/'",
        "_mirror('LO_DL_BASE', 'https://download.documentfoundation.org') + '/libreoffice/stable/'",
        "LO_DL_BASE",
    ),
    (
        r"'https://api\.github\.com/repos/'",
        "_mirror('GITHUB_API_BASE', 'https://api.github.com') + '/repos/'",
        "GITHUB_API_BASE",
    ),
    (
        r"'https://github\.com/'",
        "_mirror('GITHUB_BASE', 'https://github.com') + '/'",
        "GITHUB_BASE",
    ),
    (
        r"'https://registry\.npmjs\.org'",
        "_mirror('NPM_MIRROR', 'https://registry.npmjs.org')",
        "NPM_MIRROR",
    ),
    # NPM_MIRROR 已有 os.environ.get 兜底的位置，只把默认值换成 _mirror
    (
        r"os\.environ\.get\('NPM_MIRROR', 'https://registry\.npmjs\.org'\)",
        "_mirror('NPM_MIRROR', 'https://registry.npmjs.org')",
        "NPM_MIRROR",
    ),
]

ENV_HELPER = '''
def _mirror(name, default):
    """Download base that operator can redirect at an internal mirror.

    Recipes run inside a workspace with no internet in an intranet deployment.
    Keeping the upstream URL as the default means the same recipe still works
    on a connected host; only an explicit override changes it.
    """
    value = os.environ.get(name) or default
    return value.rstrip('/')
'''


def patch_script(text: str) -> tuple[str, list[str]]:
    """Apply every applicable patch. Returns the new text and applied env names."""
    applied: list[str] = []
    if "_mirror(" not in text:
        # 插到 import 区之后、第一个 def 之前
        marker = re.search(r"^(def |class )", text, re.M)
        if marker:
            text = text[: marker.start()] + ENV_HELPER + "\n" + text[marker.start():]
        else:
            text = ENV_HELPER + "\n" + text

    for pattern, replacement, env in PATCHES:
        new_text, count = re.subn(pattern, replacement, text)
        if count:
            text = new_text
            if env not in applied:
                applied.append(env)
    return text, applied


def main() -> int:
    if len(sys.argv) < 3:
        print("usage: patch_recipes.py <src-dir> <out-dir> [add|keep]", file=sys.stderr)
        return 2
    src, out = sys.argv[1], sys.argv[2]
    mode = sys.argv[3] if len(sys.argv) > 3 else "add"

    os.makedirs(out, exist_ok=True)
    reported = []
    for name in sorted(os.listdir(src)):
        path = os.path.join(src, name)
        if not os.path.isfile(path) or not name.endswith(".tgz"):
            continue
        dep_id = name[:-4]

        with tempfile.TemporaryDirectory() as tmp:
            subprocess.run(["tar", "-xzf", path, "-C", tmp], check=True)
            install = os.path.join(tmp, "install.sh")
            if not os.path.isfile(install):
                print(f"skip {dep_id}: no install.sh", file=sys.stderr)
                continue
            with open(install, encoding="utf-8") as fh:
                original = fh.read()

            patched, applied = patch_script(original)
            if applied:
                with open(install, "w", encoding="utf-8") as fh:
                    fh.write(patched)
                reported.append((dep_id, applied))

            # 重新打包，保持扁平文件名（校验只接受清单声明的文件）
            files = sorted(f for f in os.listdir(tmp) if os.path.isfile(os.path.join(tmp, f)))
            target = os.path.join(out, name)
            env = dict(os.environ, COPYFILE_DISABLE="1")
            subprocess.run(["tar", "-czf", target, "-C", tmp, *files], check=True, env=env)

    for dep_id, applied in reported:
        print(f"{dep_id:22} {' '.join(applied)}")
    if not reported:
        print("no recipe needed patching")
    return 0


if __name__ == "__main__":
    sys.exit(main())
