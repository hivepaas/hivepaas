import AxeBuilder from "@axe-core/playwright";

import { systemApp, systemProject } from "../../support/api";
import { expect, test } from "../../support/fixtures";

// The rules of WCAG 2.1, A and AA, that axe checks on a page.
const WCAG = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa"];

// Rules the dashboard does not pass yet, each on many pages, through shared
// components (specs/level-0.md lists them). They are left out until fixed; any
// other rule a page breaks fails it - a new kind of problem is caught.
const NOT_YET = [
    "color-contrast", // muted text on its background, below 4.5:1 - a design choice
    "button-name", // selects with no name: a table's page size, a form's options
    "aria-valid-attr-value", // tabs that switch a form's fields, with no panel to control
    "link-name", // a table row's icon links
    "label", // a few inputs named by nothing
    "nested-interactive", // a template's card, a button with buttons in it
    "svg-img-alt", // a template's icon
];

// Pages of every kind: lists, forms, settings, an app's tabs - the app being the
// HivePaaS project's Traefik, which every installation has. Nothing is written.
const PAGES = [
    "/home/",
    "/projects/",
    "/projects/new/compose/",
    "/cluster/nodes/",
    "/cluster/volumes/create/",
    "/integrations/basic-auth/create/",
    "/integrations/notification-targets/create/",
    "/operations/audit-logs/",
    "/settings/data-cleanup/configuration/",
    "/system/hivepaas/general/",
    "/user-management/users/",
    "/current-user/profile/",
    "{project}/apps/",
    "{project}/app-templates/",
    "{app}/general/",
    "{app}/deployments/",
    "{app}/env-variables/",
    "{app}/routing-settings/",
    "{app}/resources/",
    "{app}/sched-jobs/create/",
    "{app}/deployment-settings/",
];

for (const path of PAGES) {
    test(`${path} breaks no accessibility rule but those known`, async ({ page, api }) => {
        const project = await systemProject(api);
        const app = await systemApp(api, project, "traefik");
        await page.goto(
            path
                .replace("{project}", `/projects/${project.id}`)
                .replace("{app}", `/projects/${app.projectId}/${app.env}/apps/${app.id}`),
        );
        await page.waitForLoadState("networkidle");

        const results = await new AxeBuilder({ page }).withTags(WCAG).disableRules(NOT_YET).analyze();
        const broken = results.violations.map(
            v => `${v.id} (${v.impact}): ${v.help} - ${v.nodes.map(n => n.target.join(" ")).join(", ")}`,
        );
        expect(broken).toEqual([]);
    });
}
