# Test plans, by level

Each level costs more to run than the one before, and runs less often.

| Level | What                                                                                  | Where it runs                     | When                |
| ----- | ------------------------------------------------------------------------------------- | --------------------------------- | ------------------- |
| 0     | Signing in, every page opening, nothing written                                       | Anywhere, with an admin account   | Every pull request  |
| 1     | Data made, changed and removed through the screens; no container made                 | A database that is there to go    | Every pull request  |
| 2     | Deploying on one node: apps from images, functions, templates, domains, logs, volumes | A swarm that is there to go       | Nightly             |
| 3     | What needs the outside: Git and previews, Let's Encrypt, S3, email, SSO, logging, OBI | A Linux server with a real domain | Before each release |
| 4     | Clusters and the system: nodes, placement, the registry, updates, Traefik's settings  | A cluster of its own              | Before each release |

The plans: [level 0](level-0.md), [level 1](level-1.md), [level 2](level-2.md).
