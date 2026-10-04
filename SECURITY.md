# Security Policy

HivePaaS runs your apps, holds their secrets and has root on your servers through
its agent. We take reports of vulnerabilities seriously, and thank everyone who
reports one responsibly.

## Supported versions

HivePaaS is in beta. Fixes go into the **latest release** only: update to it
before reporting, and to receive a fix.

| Version | Supported |
| :--- | :--- |
| Latest `1.0.0-beta` release | ✅ |
| Older releases | ❌ |

## Reporting a vulnerability

**Do not open a public issue, discussion or pull request for a vulnerability.**

Report it privately through GitHub:
**[Report a vulnerability](https://github.com/hivepaas/hivepaas/security/advisories/new)**
(the *Security* tab of this repository, then *Report a vulnerability*).

Please include:

- **The HivePaaS version**, from the dashboard under System > HivePaaS.
- **What the vulnerability lets someone do**, and who that someone has to be:
  anyone on the internet, a signed-in user, a user of one project, an app.
- **How to reproduce it**: the steps, requests or a proof of concept.
- **Your setup**, when it matters: one server or a cluster, the OS, any proxy in
  front, such as Cloudflare.

## What happens next

- We acknowledge the report, and keep you informed while we investigate and fix.
- Once a fix is released, we publish a security advisory, and credit you in it
  unless you would rather not be named.
- Please give us a reasonable time to release a fix before you disclose the
  vulnerability publicly.

## Testing

- Test against **your own installation**.
- The public demo servers are shared and read-only: do not run load or
  denial-of-service tests against them, and do not try to reach other visitors'
  data.
