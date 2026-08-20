<div align="center">

# 🐝 HivePaaS

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

* **Instant App Deployments:** Deploy Web Apps, APIs, Databases, and Background Workers from **Docker Images** or **Git Repositories** with automated webhooks and native **GitHub App integration**.
* **Pull Request Preview Deployments:** Automatically spin up isolated ephemeral preview environments for your apps directly triggered from Pull Request comments.
* **Rich Application Management:**
  * **1-Click App Cloning:** Clone entire apps along with configurations across environments and projects.
  * **Backup & Restore:** Automated and on-demand app data backup and restoration to S3-compatible storage or local server storage.
  * **Scheduled Jobs:** Run cron-based background jobs and scheduled tasks.
  * **Health Checks:** Native container health monitoring with auto-healing and traffic draining.
* **Authentication & Custom OAuth:** Flexible sign-in supporting password-based auth, Two-Factor Authentication (2FA / TOTP), and **Custom OAuth2 / OIDC providers**.
* **Multi-Channel Event Notifications:** Real-time alerts for deployments, builds, and system events via **Email (SMTP)**, **Slack**, **Discord**, and **Telegram**.
* **Multi-User & Granular Access Control:** Team collaboration with fine-grained Role-Based Access Control (RBAC) scoped by Project and Module.
* **Automated SSL / TLS Management:** Native Let's Encrypt ACME integration (HTTP-01, DNS-01, RFC2136) with automatic background certificate renewal and custom SSL certificate support.
* **Traefik Ingress Integration:** Real-time zero-downtime hot-reloading for routing, load balancing, rate limiting, and SSL certificates without dropping active TCP connections.
* **Multi-Project & Multi-Environment:** Easily organize your apps into Projects and isolated Environments (`production`, `staging`, `development`).
* **Enterprise-grade Security:**
  * Two-Factor Authentication (2FA / TOTP) with brute-force lockout protection.
  * Argon2id state-of-the-art password hashing.
  * Layered network isolation with dedicated per-environment overlay networks.
* **Multi-Node Swarm Clustering:** Scale from a single $5 VPS to multi-node clusters with dedicated Worker nodes and flexible Placement Constraints.
* **Lightweight & High Performance:** Written in Go with minimal resource footprint (< 500MB RAM for control plane vs 2-4GB for Kubernetes).
* **Real-time Monitoring & Logs:** Stream real-time container logs and manage container lifecycles seamlessly.

---

## 🌐 Website & Demo

* **Official Website:** [https://hivepaas.com](https://hivepaas.com)
* **Demo Server:** *Coming soon*

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

* Linux VPS / Server (Ubuntu 22.04+, Debian 12+, Rocky Linux, Alpine).
* Docker Engine 24.0+ with Swarm mode support.
* Ports `80` and `443` open on host firewall.

### 1. Installation

Run the automated installer script on your primary server:

```bash
# Clone the repository
git clone https://github.com/hivepaas/hivepaas.git
cd hivepaas

# Run the local installer
./deployment/local/install.sh
```

### 2. Access the Dashboard

Once the installation finishes, access the HivePaaS dashboard:

* **URL:** `http://localhost:10000` (or `https://app.dev.localhost`)
* **Default Username:** `admin`
* **Default Password:** `abc123`

---

## 🛠️ Tech Stack

| Component | Technology | Description |
| :--- | :--- | :--- |
| **Backend Core** | [Go (Golang)](https://go.dev/) | High-performance, concurrent core engine |
| **Database** | [PostgreSQL 18](https://www.postgresql.org/) | Reliable relational storage for metadata and state |
| **Cache & Queue** | [Redis 8](https://redis.io/) | In-memory caching, rate-limiting, and session store |
| **Orchestration** | [Docker Swarm](https://docs.docker.com/engine/swarm/) | Native container clustering and scheduling |
| **Ingress Proxy** | [Traefik v3](https://traefik.io/) | Dynamic reverse proxy with automated SSL |
| **Frontend** | React, Vite, TypeScript | Modern dashboard interface |

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'feat: Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request using our [PR Template](.github/pull_request_template.md)

---

## 📄 License

Distributed under the Apache 2.0 License. See `LICENSE` for more information.
