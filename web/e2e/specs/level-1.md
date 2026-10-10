# Level 1 - data, no container

What a test needs is made through the API (`api`), named with `e2eName()` and
removed by `cleanup()`; the screens are for the flow under test. Nothing here
makes a container, but it writes: run it on an installation that is there to go.
Settings are the installation's own - those tests run one at a time and put
back what they found.

| #    | Scenario                                                                                                                           | Test                                   |
| ---- | ---------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------- |
| 1.1  | A project made from the dialog, with its two environments; a name required; a name taken                                           | `tests/level-1/projects.spec.ts`       |
| 1.1  | A project renamed and given an environment, kept after a reload; deleted after typing its name                                     | `tests/level-1/projects.spec.ts`       |
| 1.2  | A member invited by link to one project signs up, signs in, and sees that project alone                                            | `tests/level-1/users.spec.ts`          |
| 1.2  | A member disabled is turned away at sign-in                                                                                        | `tests/level-1/users.spec.ts`          |
| 1.3  | An API key shows its secret once, answers the API, and is refused once deleted                                                     | `tests/level-1/api-keys.spec.ts`       |
| 1.4  | Basic auth, key auth, SSH key, webhook: made, listed, secret masked, renamed, deleted                                              | `tests/level-1/integrations.spec.ts`   |
| 1.5  | Data Cleanup settings kept after a reload; an interval that is not one refused and not kept                                        | `tests/level-1/settings.spec.ts`       |
| 1.6  | A project made and deleted is in the audit log, with who did it                                                                    | `tests/level-1/audit-logs.spec.ts`     |
| 1.7  | An empty body on every PUT and POST is refused or harmless: never a 5xx, nothing made                                              | `tests/level-1/api-bodies.spec.ts`     |
| 1.8  | The HivePaaS project, its environments and its own apps: deleting, disabling and stopping refused                                  | `tests/level-1/system-project.spec.ts` |
| 1.9  | A read-only API key reads and is refused a write; a key without delete, a deletion                                                 | `tests/level-1/api-keys.spec.ts`       |
| 1.10 | A CLI built for an older API level reads, told the server's, and is refused its writes (426)                                       | `tests/level-1/cli-level.spec.ts`      |
| 1.11 | A member given a project to read sees its apps' settings; Save is out of reach, and the API refuses saving, deleting, making (401) | `tests/level-1/users.spec.ts`          |
| 1.12 | An API key is good until it expires, and refused from then on                                                                      | `tests/level-1/api-keys.spec.ts`       |
| 1.13 | A member turns two-factor sign-in on from the profile; then a code is asked for, a wrong one refused, the right one lets them in   | `tests/level-1/users.spec.ts`          |
| 1.14 | A member given development is refused production's apps: their address, the project's list, an address of development              | `tests/level-1/permissions.spec.ts`    |
| 1.15 | A member changes nothing an admin decides of them - role, grants, expiry - and does not delete their own account                   | `tests/level-1/permissions.spec.ts`    |

Left for later levels: Logging and Registry settings (saving them deploys) -
level 2; Traefik and HivePaaS routing and security (they can restart the proxy)

- level 4.

## Found with members and environments (2026-10-10), and fixed

- **An app was reached by the grant of the env its address named**: it was
  looked up by its project and ID alone, so a member given development read a
  production app at an address of development, saved its variables, and set
  how it clones. An app is checked by its own env now.
- **A project's list of apps had every env's**: a member given development was
  listed production's apps too. A project read through a grant on some of its
  envs lists their apps alone.
- **A member changed what an admin decides of them**: their own account's
  address - by its ID or as `current` - saved it without the Users module. The
  role and the grants have checks of their own; the expiry had none, and a
  member set theirs ten years out. Changing or deleting one's own account takes
  the Users module now; the profile and the password have routes of their own.

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

## Found by the empty-body sweep (2026-10-07), and fixed

- **130 of 193 PUT and POST endpoints panicked on `{}`** (a 500): their request
  types embed their fields' struct by pointer, which JSON left nil. The backend
  fills those before validating, so such a body is now a 400.
- **Email accounts, IM services and SSL providers were made from `{}`**,
  nameless and of no kind: they now require a name and a kind.
- **Computing a project environment's variables always panicked**: the project
  they inherit from was not loaded.
- **An SSO callback that could not complete** - an unknown provider, a missing
  state - answered 500; it answers 401.

## Found later (2026-10-07), and fixed

- **A search typed as a table appeared was dropped**: the search box's first
  debounced run was skipped, and what was typed before it restarted that run.
  The audit log listed its newest entries whatever the box said - its test
  passed only while the project's were among them. Every table's search box
  is the same component.
