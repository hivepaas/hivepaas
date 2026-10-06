# Level 1 - data, no container

What a test needs is made through the API (`api`), named with `e2eName()` and
removed by `cleanup()`; the screens are for the flow under test. Nothing here
makes a container, but it writes: run it on an installation that is there to go.
Settings are the installation's own - those tests run one at a time and put
back what they found.

| #   | Scenario                                                                                       | Test                                 |
| --- | ---------------------------------------------------------------------------------------------- | ------------------------------------ |
| 1.1 | A project made from the dialog, with its two environments; a name required; a name taken       | `tests/level-1/projects.spec.ts`     |
| 1.1 | A project renamed and given an environment, kept after a reload; deleted after typing its name | `tests/level-1/projects.spec.ts`     |
| 1.2 | A member invited by link to one project signs up, signs in, and sees that project alone        | `tests/level-1/users.spec.ts`        |
| 1.2 | A member disabled is turned away at sign-in - **fails today, marked `test.fail`**              | `tests/level-1/users.spec.ts`        |
| 1.3 | An API key shows its secret once, answers the API, and is refused once deleted                 | `tests/level-1/api-keys.spec.ts`     |
| 1.4 | Basic auth, key auth, SSH key, webhook: made, listed, secret masked, renamed, deleted          | `tests/level-1/integrations.spec.ts` |
| 1.5 | Data Cleanup settings kept after a reload; an interval that is not one refused and not kept    | `tests/level-1/settings.spec.ts`     |
| 1.6 | A project made and deleted is in the audit log, with who did it                                | `tests/level-1/audit-logs.spec.ts`   |

Left for later levels: Logging and Registry settings (saving them deploys) -
level 2; Traefik and HivePaaS routing and security (they can restart the proxy)

- level 4.

## Found while writing them (2026-10-06), and fixed

- **A disabled user still signed in**: login answered 200 with a session.
  Sign-in now refuses an account that is not active or has expired, on every
  way in (backend `createSession`).
- **An API key needed an expiration the form did not ask for**: the form now
  requires one, within a year, as the server does.
- **An unknown API key answered 412**: it answers 401 now, as no credentials do.
- **An interval or retention that is not a duration** failed with a bare "Bad
  request": the settings forms now say so at the field.
- **Accessibility**: confirmations name their dialog by their title; the copy
  buttons beside an API key's ID and secret are named; Create Project's name
  field is labelled; fields laid out with `InfoBlock` are named by their title.
