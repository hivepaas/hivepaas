# Level 0 - smoke

Nothing is written but sessions. Runs anywhere, with an admin account: an
account with less sees fewer pages.

| #   | Scenario                                                                                                   | Expected                                                                                                | Test                                   |
| --- | ---------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- | -------------------------------------- |
| 0.1 | Sign in with a username and password; an unknown user; a saved session                                     | Home, "Welcome back"; "Email or password is incorrect" on the sign-in page; Home at once                | `tests/level-0/sign-in.spec.ts`        |
| 0.2 | Open every page the sidebar reaches                                                                        | Its heading; its own address; no error toast, no console error, no API call refused (4xx/5xx)           | `tests/level-0/pages.spec.ts`          |
| 0.3 | Log out from the user menu, then open Home                                                                 | The sign-in page, both times                                                                            | `tests/level-0/session.spec.ts`        |
| 0.4 | Open an address that does not exist                                                                        | "Sorry, the page not found", with "Back to application"                                                 | `tests/level-0/session.spec.ts`        |
| 0.5 | Follow View System Apps from System › HivePaaS › Actions                                                   | The warning about changing them; the HivePaaS project's apps, traefik and worker among them             | `tests/level-0/system-apps.spec.ts`    |
| 0.6 | Order the template store A–Z, Popular, Trending, New; then search                                          | A–Z first; each order the cards the API lists for it; a search keeps the order chosen                   | `tests/level-0/template-store.spec.ts` |
| 0.7 | Open every tab and setting of an app, every page of a project - the HivePaaS project's Traefik, read only  | The app's or the project's name; its own address; no error toast, no console error, no API call refused | `tests/level-0/project-pages.spec.ts`  |
| 0.8 | Check 21 pages of every kind - lists, forms, settings, an app's tabs - with axe, against WCAG 2.1 A and AA | No rule broken but those known not passed yet                                                           | `tests/level-0/accessibility.spec.ts`  |

A page added to the dashboard is added to the list in `pages.spec.ts`.

An app's or a project's page added to the dashboard is added to the lists in
`project-pages.spec.ts`; one of a new kind, to those of `accessibility.spec.ts`.

## Accessibility not passed yet (2026-10-09)

axe finds these on the pages 0.8 checks; the test leaves them out until they
are fixed, and any other rule broken fails it. Each comes from shared
components, so one fix clears many pages.

| Rule                    | Where                                                                       | Found on (21 pages) |
| ----------------------- | --------------------------------------------------------------------------- | ------------------- |
| `color-contrast`        | muted text (`text-muted-foreground`) on its background, below 4.5:1         | 290 elements, all   |
| `button-name`           | selects with no name: a table's page size, a form's options                 | 45, 15 pages        |
| `aria-valid-attr-value` | tabs that switch a form's fields, with no panel for `aria-controls` to name | 24, 12 pages        |
| `nested-interactive`    | a template's card - a button with buttons in it; a popover in a field       | 51, 2 pages         |
| `link-name`             | a table row's icon links                                                    | 20, 4 pages         |
| `label`                 | inputs named by nothing: a create form's name, the profile's username       | 7, 4 pages          |
| `svg-img-alt`           | a template's icon                                                           | 1                   |

`color-contrast` is the palette's: changing it changes how every page looks.
