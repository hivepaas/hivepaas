# Level 2 - deploying on one node

Apps are deployed for real: images pulled, swarm services started. Run it on
a swarm that is there to go - `yarn env:up` gives one, inside dind. Each test
has a project of its own, and deleting it takes the apps and their services.

## Done

| #   | Scenario                                                                                        | Test                         |
| --- | ----------------------------------------------------------------------------------------------- | ---------------------------- |
| 2.1 | An app deployed from an image (Deployment Settings, Deploy) runs 1/1; its deployment is Done    | `tests/level-2/apps.spec.ts` |
| 2.2 | A runtime variable saved, then the app re-deployed: the container prints it, as Copy logs shows | `tests/level-2/apps.spec.ts` |
| 2.3 | Replicas set to 2 run 2/2                                                                       | `tests/level-2/apps.spec.ts` |
| 2.4 | Stopped, an app runs 0/0; started, 1/1 again                                                    | `tests/level-2/apps.spec.ts` |
| 2.5 | Deleting an app asks for its name, then removes it                                              | `tests/level-2/apps.spec.ts` |

The images: `traefik/whoami` for an app that serves, `busybox` for one that
prints what it was given. The logs are drawn on a canvas, so a test reads them
the way a person can take them: Copy logs, then the clipboard.

## Next

- **Domains and routing**: an app reached through Traefik by its domain, with
  basic auth and a redirect. Needs a real Traefik inside dind, and a way to
  reach it from the tests.
- **Persistent storage**: a volume mounted, data kept across a re-deploy.
- **Config files and secrets** mounted into the container.
- **Functions**: one made, built on its runtime, called. The runtime images
  are large: pulled once into dind.
- **App templates**: a small one installed and running.
- **Backups**: a backup repository on a volume, a backup run, a snapshot listed.
- **Scheduled and periodic jobs** that run in the container.
- **App clone**.

## Found while writing them (2026-10-07)

- **130 of 193 PUT and POST endpoints panic on a body of `{}`**: a request DTO
  embeds its fields' struct by pointer (`*AppEnvVarsBaseReq`), left nil when
  the body names none of them, and `ModifyRequest` or `Validate` reads it. The
  recovery middleware answers 500 "Unexpected error" - not a 400 naming the
  field. Not fixed yet.
- **Accessibility**: an environment variable's value field has no name, nor
  the button beside it.
