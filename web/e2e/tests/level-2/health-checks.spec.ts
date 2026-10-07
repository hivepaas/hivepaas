import { deployImage } from "../../support/api";
import { WHOAMI, appIn, deployed } from "../../support/apps";
import { test } from "../../support/fixtures";
import { expectRuns, newHealthCheck } from "../../support/health-checks";
import { domainFor, exposeAt, visit } from "../../support/routing";

test.describe.configure({ timeout: 300_000 });

// The checks are run by the backend, which in dind is on no app's network: the
// URL they call is the backend's own, or an app's domain through the proxy.
const BACKEND = "http://localhost:10100";

test("a health check runs on its interval: done when answered as asked, failed when not", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "health");
    await newHealthCheck(page, app, "e2e-ping", `${BACKEND}/api/ping`);
    await newHealthCheck(page, app, "e2e-missing", `${BACKEND}/api/no-such-path`);

    await expectRuns(page, api, app, "e2e-ping", "Done", "Failed");
    await expectRuns(page, api, app, "e2e-missing", "Failed", "Done");
});

test("a health check of an app at its own domain passes", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "health-own");
    await deployImage(api, app, WHOAMI);
    await deployed(api, app);
    // Over plain HTTP: the proxy's certificate for a .localhost name is signed
    // by no one, and the check trusts only those that are.
    const domain = domainFor("health-own");
    await exposeAt(page, app, domain, async () => {
        await page.getByRole("group", { name: "Force HTTPS" }).getByRole("checkbox").uncheck();
    });
    // Routed first: the proxy takes a few seconds to read the app's labels.
    await visit(page, `http://${domain}/health`);

    await newHealthCheck(page, app, "e2e-own", `http://${domain}/health`);

    await expectRuns(page, api, app, "e2e-own", "Done", "Failed");
});
