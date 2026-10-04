<div align="center">

<img src="docs/assets/logo.svg" alt="HivePaaS logo" width="96" height="96">

# HivePaaS

**A lightweight, self-hosted, and modern Platform-as-a-Service (PaaS) built on Docker Swarm.**

An open-source, resource-efficient alternative to Heroku, Render, and Coolify for managing and deploying applications on your own servers.

[![Go Version](https://img.shields.io/badge/Go-1.27+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Docker](https://img.shields.io/badge/Docker-Swarm-2496ED?style=flat&logo=docker)](https://docs.docker.com/engine/swarm/)
[![Traefik](https://img.shields.io/badge/Traefik-v3-24A1C1?style=flat&logo=traefik)](https://traefik.io)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

[Features](#-key-features) • [Website & Demo](#-website--demo) • [Quick Start](#-quick-start) • [Architecture](#-architecture) • [Documentation](#-documentation) • [Contributing](#-contributing)

</div>

---

## 🌟 Key Features

### 🚀 Deploy anything

* **From a Docker image, or from a Git repository** HivePaaS builds: with your Dockerfile, or one it writes for you (Go, Node.js, Bun, Deno, Python, Ruby, PHP, Java, .NET, Rust, Elixir, Next.js, Nuxt, Astro, static sites and more).
* **Git providers:** a native **GitHub App**, or access tokens and SSH keys for GitHub, GitLab, Gitea, Bitbucket and Gogs.
* **Functions:** write a handler, and HivePaaS builds and runs it - no Dockerfile, no server code. **Node.js 24** (JavaScript or TypeScript), **Bun 1**, **Python 3.13** and **Go 1.27**, called over HTTP or on a schedule.
* **App Store:** 300+ ready-made templates - databases, CMSs, analytics, monitoring, automation and more - deployed in one click.
* **Deploy on push** through the GitHub App or repository webhooks, and **pull request previews**: an isolated copy of an app for each PR, driven by `/hivepaas` commands in its comments.

### 🌐 Domains, routing & TLS

* **Traefik v3** in front of every app, with changes applied live.
* **Automatic certificates** from **Let's Encrypt**, **ZeroSSL** or **Google Trust Services**, renewed on their own; HTTP-01, or DNS-01 through a DNS provider for **wildcard** certificates; custom certificates too.
* **Per app and per domain:** redirects, path rules, force HTTPS, basic auth, allowed IPs, request size and rate limits.
* **Databases reachable from outside** over TCP, with TLS, when a client needs them.

### ⚙️ Run & scale

* **Health checks, resource limits, replicas** and **placement** on the nodes you choose.
* **Autoscaling** of apps and functions on their requests and CPU.
* **Multi-node Docker Swarm clusters:** add worker nodes as you grow; volumes on local disks, NFS or any Docker volume driver; isolated networks per project environment.
* **Scheduled jobs and workflows:** commands in containers, function calls and backups on a cron schedule, chained into sequences.
* **Live logs, a terminal into containers**, and searchable **log history**.
* **Metrics:** CPU and memory of every app; calls, failures and latency of every function; HTTP routes and outgoing calls of apps through **eBPF**, without changing their code.

### 💾 Backups & portability

* **Encrypted, deduplicated backups** (Kopia) of app data and of HivePaaS itself, to S3-compatible storage (AWS S3, Cloudflare R2, Backblaze B2, MinIO...) or a volume - on demand or scheduled, with restores.
* **Export & import** of projects, apps and settings as a bundle, secrets encrypted with a passphrase.
* **One-click cloning** of apps across environments and projects.

### 👥 Teams & security

* **Projects and environments** (`development`, `staging`, `production`...), each isolated on its own network.
* **Role-based access** per project and per module; **API keys**; an **audit log** of every change.
* **Sign-in** with a password and **two-factor authentication** (TOTP, with brute-force lockout), or **SSO** with GitHub, GitLab, Gitea, Google, Microsoft or any OpenID Connect provider. Passwords hashed with Argon2id.
* **Notifications** by Email (SMTP), Slack, Discord, Telegram and Lark.
* **Container registry** of your own (optional), and credentials for any registry, Amazon ECR included.
* **Filtered Docker API access** for the apps that need the Docker socket, limited to what they call.

### 🔌 Integrations & operations

* **REST API** with OpenAPI docs, and an **MCP server** so AI assistants can read and operate your apps.
* **Updates from the dashboard**, verified with signed releases (Ed25519 and ML-DSA-65).
* **Safe changes:** settings that could lock you out of the dashboard are applied on trial and rolled back unless you confirm them.
* **Lightweight:** written in Go; the control plane - app, agent, database, Redis and Traefik - runs in about 300 MB of memory.

---

## 🌐 Website & Demo

* **Official Website:** [https://hivepaas.com](https://hivepaas.com)
* **Demo Servers:** the addresses and the sign-in of the public demo servers are on the [website](https://hivepaas.com/#demo).

---

## 🏗️ Architecture

HivePaaS uses a clean two-tier network and node topology for maximum security and simplicity:

```text
               ┌──────────────────────────────────────┐
               │         Internet / Users             │
               └──────────────────┬───────────────────┘
                                  │ (Port 80 / 443)
                                  ▼
┌────────────────────────────────────────────────────────────────────────┐
│ PRIMARY CONTROL-PLANE (Manager Node)                                   │
│                                                                        │
│  ┌─────────────────┐       ┌────────────────┐       ┌───────────────┐  │
│  │  Traefik Proxy  │◄─────►│  HivePaaS App  │◄─────►│  PostgreSQL   │  │
│  └────────┬────────┘       └───────┬────────┘       └───────────────┘  │
│           │                        │                                   │
└───────────┼────────────────────────┼───────────────────────────────────┘
            │                        │ (gRPC Management)
    (hivepaas_net overlay)           ▼
┌───────────┼────────────────────────────────────────────────────────────┐
│ WORKER NODES (Multi-Node Cluster)                                      │
│           │                                                            │
│           ├────────────────────────┬────────────────────────┐          │
│           ▼                        ▼                        ▼          │
│  ┌─────────────────┐      ┌─────────────────┐      ┌────────────────┐  │
│  │   Web App (A)   │      │   Web App (B)   │      │ HivePaaS Agent │  │
│  │  (project_net)  │      │  (project_net)  │      │  (Global Mode) │  │
│  └─────────────────┘      └─────────────────┘      └────────────────┘  │
└────────────────────────────────────────────────────────────────────────┘
```

* **`hivepaas_net`:** Shared Overlay network for Traefik to route ingress traffic to publicly exposed containers.
* **`project_env_net`:** Completely isolated private overlay networks for internal communication (e.g. App to Database/Redis).

---

## 🚀 Quick Start

### Prerequisites

* A Linux server with root access: Debian, Ubuntu, Fedora, RHEL, Rocky, AlmaLinux, Amazon Linux, SLES, openSUSE, Arch or Alpine.
* 4 CPUs, 8 GB of memory and 40 GB of disk recommended; HivePaaS runs on less.
* Ports `80` and `443` free and open to the internet.
* Docker 29.5 or newer - the installer installs or upgrades it if needed.

### 1. Install

On the server:

```bash
curl -fsSL https://get.hivepaas.com | sudo bash
```

The installer asks for the admin's email and password and the dashboard's domain, sets up Docker Swarm, and deploys HivePaaS. To install without questions, see [Silent install](https://docs.hivepaas.com/docs/installation/silent-install).

### 2. Open the dashboard

Open the domain you gave, such as `https://hivepaas.example.com`, and sign in with the admin's email and password.

### 3. If a change locks you out

Configuration changes that can make the dashboard unreachable - traefik's startup
command, the HivePaaS routing and proxy settings - are applied on trial and undone
automatically unless you confirm them. [docs/recovery.md](docs/recovery.md)
explains what catches what, and what to do by hand when nothing automatic can run.

To work on HivePaaS itself, see [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).

---

## 🛠️ Tech Stack

| Component | Technology | Required? | What it does |
| :--- | :--- | :--- | :--- |
| **Backend** | [Go](https://go.dev/) ([Gin](https://gin-gonic.com/), [Bun](https://bun.uptrace.dev/), [lego](https://go-acme.github.io/lego/)) | Required | The API, the task queue and the dashboard's server |
| **Agent** | Go, [gRPC](https://grpc.io/) | Required | Runs on every node: builds images, runs commands and backups, reads the node |
| **Database** | [PostgreSQL 18](https://www.postgresql.org/) | Required | HivePaaS's state: projects, apps, settings, tasks |
| **Cache & queue** | [Redis 8](https://redis.io/) | Required | Sessions, locks, the task queue and rate limiting |
| **Orchestration** | [Docker Swarm](https://docs.docker.com/engine/swarm/) (Docker 29.5+) | Required | Runs and schedules every container, on one node or many |
| **Ingress** | [Traefik v3](https://traefik.io/) | Required | Routing, TLS and the per-app rules |
| **Image builds** | [BuildKit](https://docs.docker.com/build/buildkit/) (docker buildx) | Required to build from Git and functions | Builds images from repositories and functions' code |
| **Backups** | [Kopia](https://kopia.io/) | Optional - when you back up | Encrypted, deduplicated backups and restores |
| **Logs & metrics** | [VictoriaLogs](https://docs.victoriametrics.com/victorialogs/) and vlagent | Optional - switched on in System › Logging | Log history, metrics and autoscaling |
| **Routes & calls** | [OBI](https://opentelemetry.io/docs/zero-code/obi/) (OpenTelemetry eBPF Instrumentation) | Optional - per app | HTTP routes and outgoing calls of apps, without code changes |
| **Registry** | [zot](https://zotregistry.dev/) | Optional - switched on in System › Registry | A container registry of your own, for images built on a multi-node cluster |
| **Function runtimes** | [hivepaas/function-runtimes](https://github.com/hivepaas/function-runtimes) | Optional - when you use functions | The images functions are built on |
| **Dashboard** | [React 19](https://react.dev/), [Vite](https://vite.dev/), TypeScript, Tailwind CSS, TanStack Query | Required | The web interface |

---

## 📚 Documentation

* **User documentation:** [docs.hivepaas.com](https://docs.hivepaas.com) - installation, deploying apps, domains, backups, clusters and troubleshooting.
* **REST API:** the [API reference](https://docs.hivepaas.com/api/hivepaas-api), from the OpenAPI description in [docs/openapi/swagger.json](docs/openapi/swagger.json).
* **Developing HivePaaS:** [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) and [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).
* **Releasing:** [docs/RELEASING.md](docs/RELEASING.md).

---

## 🗺️ Roadmap

Planned, in no particular order and without dates:

* **CLI** - deploy, follow logs and run jobs from your terminal and your CI.
* **Functions** - API keys to protect a function's endpoint, and bring-your-own runtime images.
* **Backups** - network volumes, such as NFS, as backup repositories.
* **Registry** - read-only accounts, and retention rules per repository.

Ideas and requests are welcome in [GitHub Issues](https://github.com/hivepaas/hivepaas/issues).

---

## 🔒 Security

Please report vulnerabilities **privately**, through
[GitHub's vulnerability reporting](https://github.com/hivepaas/hivepaas/security/advisories/new) -
never in a public issue. See [SECURITY.md](SECURITY.md) for what to include and what happens next.

---

## 💬 Community & Support

* **[Discord](https://discord.com/invite/2TgD3zDb2e)** - questions, and help from the team and the community.
* **[GitHub Issues](https://github.com/hivepaas/hivepaas/issues)** - bugs and feature requests.
* **[Getting help](https://docs.hivepaas.com/docs/troubleshooting/getting-help)** - what to include so a question gets an answer sooner.

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!

Setting up a development machine - the local cluster, the three ways to run the
backend, and what to run before you push - is in
[docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'feat: Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request using our [PR Template](.github/pull_request_template.md)

---

## 📄 License

Distributed under the Apache 2.0 License. See `LICENSE` for more information.
