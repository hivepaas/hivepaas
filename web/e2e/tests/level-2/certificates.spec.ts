import type { APIRequestContext, Page } from "@playwright/test";
import { execFileSync } from "node:child_process";
import crypto from "node:crypto";
import fs from "node:fs";
import https from "node:https";
import tls from "node:tls";

import { type App, createApp, createProject, deleteProject, deployImage } from "../../support/api";
import { WHOAMI, deployed } from "../../support/apps";
import { env } from "../../support/env";
import { type Cleanup, e2eName, expect, test } from "../../support/fixtures";
import { domainFor, exposeAt } from "../../support/routing";

test.describe.configure({ timeout: 300_000 });

interface KeyPair {
    certificate: string;
    privateKey: string;
}

// selfSigned is a certificate of one's own for the domain, as a person brings
// one: made by openssl, for 30 days.
function selfSigned(dir: string, domain: string): KeyPair {
    const key = `${dir}/${domain}.key`;
    const cert = `${dir}/${domain}.crt`;
    execFileSync(
        "openssl",
        [
            "req",
            "-x509",
            "-newkey",
            "rsa:2048",
            "-nodes",
            "-days",
            "30",
            "-keyout",
            key,
            "-out",
            cert,
            "-subj",
            `/CN=${domain}`,
            "-addext",
            `subjectAltName=DNS:${domain}`,
        ],
        { stdio: "ignore" },
    );
    return { certificate: fs.readFileSync(cert, "utf8"), privateKey: fs.readFileSync(key, "utf8") };
}

// servedBy reports whether the proxy serves the domain with the certificate:
// a TLS connection trusting it alone, and checking the name, succeeds only
// then. The answer is the served certificate's fingerprint, or "" when not.
async function servedBy(domain: string, certificate: string): Promise<string> {
    const [host = "127.0.0.1", port = "10443"] = env.ingressHTTPS.split(":");
    return new Promise(resolve => {
        const socket = tls.connect({ host, port: Number(port), servername: domain, ca: [certificate] }, () => {
            const fingerprint = socket.getPeerX509Certificate()?.fingerprint256 ?? "";
            socket.end();
            resolve(fingerprint);
        });
        socket.on("error", () => {
            resolve("");
        });
    });
}

// getTrusting asks the domain for / over HTTPS, at the proxy, trusting the
// certificate alone, and answers the status and the body.
async function getTrusting(domain: string, certificate: string): Promise<{ status: number; body: string }> {
    const [host = "127.0.0.1", port = "10443"] = env.ingressHTTPS.split(":");
    return new Promise((resolve, reject) => {
        const req = https.get(
            { host, port: Number(port), path: "/", servername: domain, headers: { Host: domain }, ca: [certificate] },
            res => {
                let body = "";
                res.on("data", (chunk: Buffer) => (body += chunk.toString()));
                res.on("end", () => {
                    resolve({ status: res.statusCode ?? 0, body });
                });
            },
        );
        req.on("error", reject);
    });
}

// whoamiNamed is a project of its own with a whoami app named after the test:
// two apps of one name, in any two projects, would share the proxy's file of
// the app's configuration (see the spec's findings).
async function whoamiNamed(api: APIRequestContext, cleanup: Cleanup, label: string): Promise<App> {
    const project = await createProject(api, e2eName(label));
    cleanup(() => deleteProject(api, project.id));
    const app = await createApp(api, project, label);
    await deployImage(api, app, WHOAMI);
    await deployed(api, app);
    return app;
}

// newCustomCertificate pastes a certificate and its key into a project's SSL
// Certificates, as Custom, and saves it.
async function newCustomCertificate(page: Page, projectId: string, domain: string, pair: KeyPair): Promise<void> {
    await page.goto(`/projects/${projectId}/integrations/ssl-certificates/create/`);
    await page.getByRole("group", { name: "Domain *" }).getByRole("textbox").fill(domain);
    await page.getByRole("group", { name: "Certificate Type *" }).getByRole("combobox").click();
    await page.getByRole("option", { name: "Custom", exact: true }).click();
    await page.getByRole("group", { name: "Certificate *" }).getByRole("textbox").fill(pair.certificate);
    await page.getByRole("group", { name: "Private Key *" }).getByRole("textbox").fill(pair.privateKey);
    await page.getByRole("button", { name: "Save" }).click();
}

test("a certificate of one's own, pasted in, is the one its domain is served with", async ({
    page,
    api,
    cleanup,
}, testInfo) => {
    const app = await whoamiNamed(api, cleanup, "own-cert");
    const domain = domainFor("own-cert");
    fs.mkdirSync(testInfo.outputDir, { recursive: true });
    const pair = selfSigned(testInfo.outputDir, domain);

    await newCustomCertificate(page, app.projectId, domain, pair);
    // It expires when it says: read from it, not typed in beside it.
    const expiry = await page.evaluate(
        iso => {
            const at = new Date(iso);
            const two = (n: number) => String(n).padStart(2, "0");
            return `${at.getFullYear()}-${two(at.getMonth() + 1)}-${two(at.getDate())} ${two(at.getHours())}:${two(at.getMinutes())}`;
        },
        new Date(new crypto.X509Certificate(pair.certificate).validTo).toISOString(),
    );
    await expect(page.getByRole("row", { name: new RegExp(domain) })).toContainText(expiry);
    await exposeAt(page, app, domain, async () => {
        await page.getByRole("group", { name: "SSL Certificate" }).getByRole("combobox").click();
        await page.getByRole("option", { name: domain, exact: true }).click();
    });

    const fingerprint = new crypto.X509Certificate(pair.certificate).fingerprint256;
    await expect
        .poll(() => servedBy(domain, pair.certificate), { timeout: 120_000, intervals: [2_000] })
        .toBe(fingerprint);
    // The app answers through it, once the proxy routes the domain: whoami
    // tells the request it got.
    let answer = { status: 0, body: "" };
    await expect
        .poll(
            async () => {
                answer = await getTrusting(domain, pair.certificate);
                return answer.status;
            },
            { timeout: 120_000, intervals: [2_000] },
        )
        .toBe(200);
    expect(answer.body).toContain(`Host: ${domain}`);
});

// A certificate that cannot serve its domain - the private key of another -
// is refused, and says why, rather than saved: the proxy would serve the
// domain with its default certificate, which no browser trusts.
test("a certificate pasted with the private key of another is refused", async ({ page, api, cleanup }, testInfo) => {
    const app = await whoamiNamed(api, cleanup, "wrong-key");
    const domain = domainFor("wrong-key");
    fs.mkdirSync(testInfo.outputDir, { recursive: true });
    const pair = selfSigned(testInfo.outputDir, domain);
    const other = selfSigned(testInfo.outputDir, domainFor("other-key"));

    await newCustomCertificate(page, app.projectId, domain, {
        certificate: pair.certificate,
        privateKey: other.privateKey,
    });
    await expect(
        page.getByText("The certificate cannot be used: the private key is not the certificate's"),
    ).toBeVisible();
    await page.goto(`/projects/${app.projectId}/integrations/ssl-certificates/`);
    await expect(page.getByRole("row", { name: new RegExp(domain) })).toHaveCount(0);
});
