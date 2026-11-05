#!/bin/sh
# 重新生成每个依赖的 artifact.tar.gz。
#
# artifact.tar.gz 是构建产物（由 dependency.yaml 和脚本打出来），不进版本库。
# 发布前或 clone 之后跑一次这个脚本。
#
# COPYFILE_DISABLE=1 是必须的：macOS 的 tar 默认会把 AppleDouble 的 ._xxx
# 元数据文件一起打进包，而校验只接受清单里声明的文件，多一个就整个依赖
# 拒绝加载。这个坑在 Linux 上不会出现，在 Mac 上必现。
set -eu

cd "$(dirname "$0")/dependencies"

for dir in */; do
	id=${dir%/}
	[ -f "$id/dependency.yaml" ] || continue

	# 归档内容以清单为准：清单声明了哪个脚本，就打哪个。
	# 这里不猜文件名，避免打包出一个校验不过的包。
	files="dependency.yaml"
	for script in install.sh update.sh remove.sh reinstall.sh check-update.sh version.sh; do
		[ -f "$id/$script" ] && files="$files $script"
	done
	# 图标：只在清单里出现 icon: 时才打进去。
	if grep -qE '^icon:[[:space:]]*[^[:space:]]' "$id/dependency.yaml"; then
		icon=$(sed -n 's/^icon:[[:space:]]*//p' "$id/dependency.yaml" | head -1 | tr -d '"'"'"' ')
		[ -f "$id/$icon" ] && files="$files $icon"
	fi

	echo "packing $id: $files"
	# shellcheck disable=SC2086
	(cd "$id" && COPYFILE_DISABLE=1 tar -czf artifact.tar.gz $files)
done

echo "done"
