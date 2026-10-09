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

axe finds these on the pages 0.8 checks; the test leaves them out, and any
other rule broken fails it.

- `color-contrast`: muted text (`text-muted-foreground`) on its background,
  below 4.5:1 - 290 elements, on every page. The palette's: changing it
  changes how every page looks.
- `nested-interactive`: a date picker's Clear date is a control inside the
  button that opens the calendar. Moved out beside it, waiting for review: it
  is where the eye sees it.

## Found with axe (2026-10-09), and fixed

- **Controls named by nothing** (`button-name`, `label`): a checkbox, a select
  or an input alone in a block titled "Enabled", "Default", "Name" had no
  name. A block's title now names what it holds that nothing else names - a
  select by its title and its value, "Environment development" - and a
  table's page size is "Rows per page".
- **Tabs with no panel** (`aria-valid-attr-value`): tabs that switch a form's
  fields, or a page's sections by their routes, said they controlled panels
  that were not there. Such a tab names none now.
- **A table row's view link** (`link-name`) was an eye alone: "View app",
  "View project", "View node", "View user".
- **A template's card was a button with buttons in it** (`nested-interactive`):
  it is a group named by the template; a click on it still opens the
  details, and the keyboard has the Details button - Enter on Deploy opened
  the details before, the card taking the key.
- **The template's GitHub icon** (`svg-img-alt`) is hidden from screen readers,
  beside the link's text.
