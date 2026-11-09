<div align="center">
  <img src="./assets/logo.png" alt="Ami" height="80">
  <h1>Ami</h1>
  <p>东风集团 · AI 协作平台<br>
  每位员工一个专属智能体 —— 独立工作区、长期记忆、内部系统接入，全部部署在内网。</p>
</div>

## 这是什么

Ami 是东风集团的 AI 协作平台。每位员工拥有且仅有一个智能体（东风助手），
配有独立的工作区、文件系统、浏览器与长期记忆；专家（Expert）与技能（Skill）
作为子角色服务于这个智能体，而不是需要单独管理的实体。

平台完全私有化部署在集团内网，不依赖任何外部 SaaS 服务，也不对外发送任何数据。

## 主要能力

- **独立工作区**：每个智能体有自己的文件系统、网络与浏览器，任务与文件互不干扰
- **长期记忆**：跨会话记住用户的偏好、习惯与历史决策
- **专家与技能**：为智能体指派专家角色，并从技能库中挑选它可用的能力
- **定时任务**：按日程自动执行例行工作，结果通过站内消息送达
- **内部模型接入**：对接擎天 AI 中台的模型服务，全部走内网

## 部署

平台以 Docker Compose 方式部署在内网服务器上。详细步骤与环境变量见
[DEPLOYMENT.md](DEPLOYMENT.md)。

```bash
cp conf/app.docker.toml config.toml
# 按内网环境编辑 config.toml
export AMI_INTERNAL_RPC_SHARED_SECRET="$(openssl rand -hex 32)"
docker compose up -d
```

Compose 会分别启动 Server 与 Channel 服务。请妥善保存内部 RPC 密钥，
重新创建服务时继续使用同一个值。

拆分部署中，Channel 通过认证后的内部 Storage RPC 流式读写媒体，
只有 Server 需要访问工作区和媒体存储。升级时先更新 Server，再更新 Channel。

不使用 Docker（或从已有裸机部署升级）？在 `config.toml` 中将
`internal_rpc.shared_secret` 留空即可：Server 会内嵌 Channel 运行时，
继续以单进程方式运行。

## 目录结构

| 路径 | 说明 |
| --- | --- |
| `apps/web` | 前端（Vue 3 + Vite） |
| `apps/desktop` | 桌面客户端（Electron） |
| `internal/` | 后端服务（Go） |
| `packages/ui` | 设计系统与组件库 |
| `packages/icons` | 图标包 |
| `skillhub/` | 内网技能仓库 |

## 许可证

AGPLv3
