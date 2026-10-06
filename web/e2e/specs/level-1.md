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

## Found while writing them (2026-10-06)

- **A disabled user still signs in**: `login-with-password` answers 200 with a
  session, and only the calls after it say "User is unavailable". The test is
  marked `test.fail`; remove the mark once sign-in refuses them.
- **An API key needs an expiration the form does not ask for**: Create Key
  without a date fails with "Param 'expireAt': Value is required".
- **An unknown API key answers 412** (`ERR_API_KEY_INVALID`), where no
  credentials answer 401.
- **A Data Cleanup interval that is not one** fails with a bare "Bad request",
  naming no field.
- **Accessibility**: the delete confirmations have no accessible name (found by
  their heading); the copy buttons beside an API key's ID and secret have no
  name; the Create Project dialog's name field is named only by its
  placeholder. Fields laid out with `InfoBlock` had no name either: fixed with
  these tests, the block is now a group named by its title.
