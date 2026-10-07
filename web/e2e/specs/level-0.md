# Level 0 - smoke

Nothing is written but sessions. Runs anywhere, with an admin account: an
account with less sees fewer pages.

| #   | Scenario                                                               | Expected                                                                                      | Test                                |
| --- | ---------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- | ----------------------------------- |
| 0.1 | Sign in with a username and password; an unknown user; a saved session | Home, "Welcome back"; "Email or password is incorrect" on the sign-in page; Home at once      | `tests/level-0/sign-in.spec.ts`     |
| 0.2 | Open every page the sidebar reaches                                    | Its heading; its own address; no error toast, no console error, no API call refused (4xx/5xx) | `tests/level-0/pages.spec.ts`       |
| 0.3 | Log out from the user menu, then open Home                             | The sign-in page, both times                                                                  | `tests/level-0/session.spec.ts`     |
| 0.4 | Open an address that does not exist                                    | "Sorry, the page not found", with "Back to application"                                       | `tests/level-0/session.spec.ts`     |
| 0.5 | Follow View System Apps from System › HivePaaS › Actions               | The warning about changing them; the HivePaaS project's apps, traefik and worker among them   | `tests/level-0/system-apps.spec.ts` |

A page added to the dashboard is added to the list in `pages.spec.ts`.
