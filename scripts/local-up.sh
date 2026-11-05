#!/usr/bin/env bash
# 用本地源码构建的 ami 镜像启动整套服务。
# 所有基础镜像与镜像构建都走本机 7890 代理，因为默认网络出口不通。
set -euo pipefail
cd "$(dirname "$0")/.."

PROXY=${PROXY:-http://host.docker.internal:7890}
PROXY_ARGS="-f docker/Dockerfile.server"

if [ ! -f .env.runtime ]; then
  echo "AMI_INTERNAL_RPC_SHARED_SECRET=$(openssl rand -hex 32)" > .env.runtime
  echo "POSTGRES_PASSWORD=ami123" >> .env.runtime
  echo "已生成 .env.runtime"
fi

if [ "${SKIP_BUILD:-0}" != "1" ]; then
  echo "==> 构建 ami/server:local"
  docker build -f docker/Dockerfile.server -t ami/server:local --build-arg VERSION=dev \
    --build-arg HTTP_PROXY=$PROXY --build-arg HTTPS_PROXY=$PROXY \
    --build-arg http_proxy=$PROXY --build-arg https_proxy=$PROXY \
    --network host .
  echo "==> 构建 ami/web:local"
  docker build -f docker/Dockerfile.web -t ami/web:local \
    --build-arg HTTP_PROXY=$PROXY --build-arg HTTPS_PROXY=$PROXY \
    --build-arg http_proxy=$PROXY --build-arg https_proxy=$PROXY \
    --network host .
fi

echo "==> 启动"
docker compose --env-file .env.runtime \
  -f docker-compose.yml -f docker-compose.local.yaml up -d

echo
echo "  Web UI : http://localhost:8082"
echo "  API    : http://localhost:8082/api"
