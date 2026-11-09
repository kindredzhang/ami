<div align="center">
  <img src="./assets/logo.png" alt="Ami" height="80">
  <h1>Ami</h1>
  <p>Dongfeng Group · AI Collaboration Platform<br>
  One dedicated assistant per employee — isolated workspace, long-term memory and internal system access, fully self-hosted.</p>
</div>

## What this is

Ami is the AI collaboration platform for Dongfeng Group. Every employee has exactly one
assistant (东风助手) with its own workspace, filesystem, browser and long-term memory.
Experts and skills are sub-roles serving that assistant, not separate entities to manage.

The platform is deployed entirely on the corporate intranet. It depends on no external
SaaS service and sends no data outside the network.

## Capabilities

- **Isolated workspace**: each assistant has its own filesystem, network and browser
- **Long-term memory**: preferences, habits and past decisions carry across sessions
- **Experts and skills**: assign expert roles, then pick which skills each one may use
- **Scheduled tasks**: routine work runs on a schedule and reports back in-app
- **Internal model access**: all inference goes through the internal model gateway

## Deployment

Deployed with Docker Compose on an intranet server. See [DEPLOYMENT.md](DEPLOYMENT.md)
for the full steps and environment variables.

```bash
cp conf/app.docker.toml config.toml
# edit config.toml for your environment
export AMI_INTERNAL_RPC_SHARED_SECRET="$(openssl rand -hex 32)"
docker compose up -d
```

The Compose stack runs Server and Channel as separate services. Keep the internal RPC
secret private and reuse the same value whenever the stack is recreated.

In split deployments, Channel streams media through authenticated internal Storage RPC.
Only Server needs access to the workspace and media storage. Upgrade Server before Channel.

Running without Docker (or upgrading an existing bare-metal install)? Leave
`internal_rpc.shared_secret` empty in `config.toml`: the server then embeds the channel
runtime and keeps running as a single all-in-one process.

## Layout

| Path | Contents |
| --- | --- |
| `apps/web` | Web front end (Vue 3 + Vite) |
| `apps/desktop` | Desktop client (Electron) |
| `internal/` | Backend services (Go) |
| `packages/ui` | Design system and component library |
| `packages/icons` | Icon set |
| `skillhub/` | Internal skill registry |

## License

AGPLv3
