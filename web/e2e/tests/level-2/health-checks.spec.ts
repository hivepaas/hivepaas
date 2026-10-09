import { deployImage } from "../../support/api";
import { WHOAMI, appIn, appPage, deployed } from "../../support/apps";
import { expect, test } from "../../support/fixtures";
import { checkRuns, expectRuns, newHealthCheck, setHealth } from "../../support/health-checks";
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

// A check reads the answer's body as asked: the backend's /api/ping answers
// {"message": "pong"}, which a text matches by a pattern, and a JSON contains;
// another text, or another JSON exactly, fails.
test("a health check reads the answer's body: a text by its pattern, a JSON by what it contains", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "health-body");
    const ping = `${BACKEND}/api/ping`;
    const body = (tab: "Text" | "JSON", field: string, value: string) => async () => {
        await page.getByRole("group", { name: "Return Body Must Be" }).getByRole("tab", { name: tab }).click();
        await page.getByRole("group", { name: field }).getByRole("textbox").fill(value);
    };
    await newHealthCheck(page, app, "e2e-text-ok", ping, body("Text", "Text Regex", "p[aeiou]ng"));
    await newHealthCheck(page, app, "e2e-text-no", ping, body("Text", "Text Exact", "ping"));
    await newHealthCheck(page, app, "e2e-json-ok", ping, body("JSON", "JSON Contains", '{"message": "pong"}'));
    await newHealthCheck(page, app, "e2e-json-no", ping, body("JSON", "JSON Exact", '{"message": "ping"}'));

    await expectRuns(page, api, app, "e2e-text-ok", "Done", "Failed");
    await expectRuns(page, api, app, "e2e-text-no", "Failed", "Done");
    await expectRuns(page, api, app, "e2e-json-ok", "Done", "Failed");
    await expectRuns(page, api, app, "e2e-json-no", "Failed", "Done");
});

// A check whose answer does not come fails at its timeout - the interval, when
// none is set - and holds no other check: the app's other one keeps running,
// and sees the app go down and up again. A run is kept when the result
// changes. A check turned off runs no more: a change goes unseen.
test("a check left waiting fails at its timeout and holds no other; one turned off runs no more", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "health-wait");
    await deployImage(api, app, WHOAMI);
    await deployed(api, app);
    const domain = domainFor("health-wait");
    await exposeAt(page, app, domain, async () => {
        await page.getByRole("group", { name: "Force HTTPS" }).getByRole("checkbox").uncheck();
    });
    await visit(page, `http://${domain}/health`);

    // whoami answers after the wait it is asked for: two minutes.
    await newHealthCheck(page, app, "e2e-hangs", `http://${domain}/?wait=120s`);
    await newHealthCheck(page, app, "e2e-meanwhile", `http://${domain}/health`);
    await expectRuns(page, api, app, "e2e-hangs", "Failed", "Done");
    await expectRuns(page, api, app, "e2e-meanwhile", "Done", "Failed");
    const latest = async () => (await checkRuns(api, app, "e2e-meanwhile"))[0];
    const RESULT = { timeout: 60_000, intervals: [3_000] };
    await setHealth(page, domain, 500);
    await expect.poll(latest, RESULT).toBe("failed");
    await setHealth(page, domain, 200);
    await expect.poll(latest, RESULT).toBe("done");

    await page.goto(appPage(app, "periodic-jobs"));
    await page
        .getByRole("row", { name: /e2e-meanwhile/ })
        .getByRole("button", { name: "Actions menu" })
        .click();
    await page.getByRole("button", { name: "Change Status" }).click();
    const dialog = page.getByRole("dialog", { name: "Change status" });
    await dialog.getByRole("tab", { name: "Disabled" }).click();
    await dialog.getByRole("button", { name: "Save" }).click();
    await expect(dialog).toBeHidden();
    // A run under way when it was turned off ends; none starts after.
    await page.waitForTimeout(12_000);
    const off = (await checkRuns(api, app, "e2e-meanwhile")).length;
    await setHealth(page, domain, 500);
    await page.waitForTimeout(30_000);
    expect((await checkRuns(api, app, "e2e-meanwhile")).length, "the change goes unseen").toBe(off);
});
