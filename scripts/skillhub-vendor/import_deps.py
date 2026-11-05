#!/usr/bin/env python3
"""把上游依赖包导入 skillhub 数据目录。

输入是已打好补丁的 .tgz（见 patch_recipes.py），输出是 skillhub 认识的结构：

    skillhub-data/dependencies/<id>/
        dependency.yaml
        artifact.tar.gz      (扁平文件名，含清单本身)
        icon.svg
        install.sh / remove.sh / check-update.sh

为什么要重新打包而不是直接复制上游的 tgz：上游包里的 install.sh 是硬编码
公网地址的版本，这里必须换成打过补丁的。其余文件原样保留。
"""
from __future__ import annotations

import os
import shutil
import subprocess
import sys
import tarfile
import tempfile
from pathlib import Path


def import_one(tgz: Path, out_root: Path) -> str:
    dep_id = tgz.stem
    with tempfile.TemporaryDirectory() as tmp:
        tmp_path = Path(tmp)
        with tarfile.open(tgz) as tar:
            # 只取扁平文件名，和校验器的要求一致
            for member in tar.getmembers():
                if not member.isfile():
                    continue
                name = os.path.basename(member.name)
                if not name or name.startswith("._"):
                    continue
                src = tar.extractfile(member)
                if src is None:
                    continue
                (tmp_path / name).write_bytes(src.read())

        manifest = tmp_path / "dependency.yaml"
        if not manifest.is_file():
            raise SystemExit(f"{dep_id}: 包里没有 dependency.yaml")

        target = out_root / dep_id
        if target.exists():
            shutil.rmtree(target)
        target.mkdir(parents=True)

        files = sorted(p.name for p in tmp_path.iterdir() if p.is_file())
        for name in files:
            shutil.copy2(tmp_path / name, target / name)

        # 归档内容必须和清单声明的一致，多一个文件就整个依赖被拒。
        # 这里直接全打，因为上面已经把 ._ 之类过滤掉了。
        env = dict(os.environ, COPYFILE_DISABLE="1")
        subprocess.run(
            ["tar", "-czf", "artifact.tar.gz", *files],
            cwd=target,
            check=True,
            env=env,
        )
        return f"{dep_id:22} {len(files)} files"


def main() -> int:
    if len(sys.argv) < 3:
        print("usage: import_deps.py <patched-dir> <skillhub-data-dir> [id...]", file=sys.stderr)
        return 2

    src = Path(sys.argv[1])
    out_root = Path(sys.argv[2]) / "dependencies"
    out_root.mkdir(parents=True, exist_ok=True)

    wanted = sys.argv[3:]
    tarballs = sorted(src.glob("*.tgz"))
    if wanted:
        tarballs = [t for t in tarballs if t.stem in wanted]

    for tgz in tarballs:
        try:
            print(import_one(tgz, out_root))
        except SystemExit as exc:
            print(f"skip {tgz.stem}: {exc}", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
