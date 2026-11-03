import { type APIRequestContext, expect } from "@playwright/test";
import { execFileSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";

import { ok } from "./api";
import { backendDir } from "./env";

// License keys for the throwaway installation, signed as env/up.sh signs its
// Colony key: by the backend's tools/licensesign, which env/up.sh builds into
// env/.build, with the development keys that only a HivePaaS in development
// trusts. A key is put in force through the API, as System › License does it.
//
// Every test but the license project's counts on that Colony key: a test that
// installs another puts it back after (restoreColony).

const signer = path.join(import.meta.dirname, "..", "env", ".build", "licensesign");
const devKeys = path.join(backendDir, "tools", "licensesign", "devkeys");
const edKey = path.join(devKeys, "dev-ed.key");
const mlKey = path.join(devKeys, "dev-ml.key");

// canSignLicenses is whether keys can be signed here: the signer built and the
// development keys found.
export const canSignLicenses = [signer, edKey, mlKey].every(file => fs.existsSync(file));

// noSigner says why keys cannot be signed, for a test that skips without them.
export const noSigner =
    `no license signer at ${signer}, or no development keys at ${devKeys}: ` +
    "env/up.sh builds the signer, from the backend repo beside this one or HP_BACKEND_DIR";

export interface LicenseUsage {
    limit: string;
    // null: no limit.
    max: number | null;
    used: number;
}

export interface LicensePortal {
    url: string;
    pricingUrl: string;
    trialUrl: string;
    accountUrl: string;
}

// License is what GET /system/license answers: the plan in force and its
// status, the stored key's own fields, the usage of each limit, and the
// portal's pages for this installation.
export interface License {
    status: string;
    plan: string;
    keyPlan?: string;
    trial: boolean;
    development: boolean;
    licenseId?: string;
    customerName?: string;
    installationId: string;
    installations?: string[];
    issuedAt?: string;
    expiresAt?: string;
    graceEndsAt?: string;
    movedAt?: string;
    reason?: string;
    limits: LicenseUsage[];
    portal?: LicensePortal;
    refresh?: { enabled: boolean; lastCheckedAt?: string; lastResult?: string; lastError?: string };
}

// readLicense is the license in force.
export async function readLicense(api: APIRequestContext): Promise<License> {
    const body = (await ok(await api.get("system/license"), "reading the license")) as { data: License };
    return body.data;
}

// installLicense puts a key in force, as Install key does, and answers the
// license it puts in force. A key refused fails the test.
export async function installLicense(api: APIRequestContext, key: string): Promise<License> {
    const body = (await ok(await api.put("system/license", { data: { key } }), "installing a license key")) as {
        data: License;
    };
    return body.data;
}

// usageOf is a limit's usage in a license.
export function usageOf(license: License, limit: string): LicenseUsage {
    const usage = license.limits.find(u => u.limit === limit);
    expect(usage, `the license's usage of ${limit}`).toBeDefined();
    return usage as LicenseUsage;
}

type PaidPlan = "comb" | "hive" | "colony" | "apiary";

export interface IssueOptions {
    plan: PaidPlan;
    installation: string;
    // Days from now; the signer's default is 30.
    days?: number;
    // A trial has no grace.
    trial?: boolean;
    nodes?: number;
    users?: number;
    licenseId?: string;
    customerName?: string;
    customerEmail?: string;
}

// issueKey signs a key of a plan, from now, as `licensesign issue` does: the
// plan's own limits unless nodes and users say otherwise.
export function issueKey(options: IssueOptions): string {
    const args = ["issue", "-key", edKey, "-key", mlKey, "-plan", options.plan, "-installation", options.installation];
    const flags: [string, string | number | undefined][] = [
        ["-days", options.days],
        ["-nodes", options.nodes],
        ["-users", options.users],
        ["-license-id", options.licenseId],
        ["-customer-name", options.customerName],
        ["-customer-email", options.customerEmail],
    ];
    for (const [flag, value] of flags) {
        if (value !== undefined) args.push(flag, String(value));
    }
    if (options.trial) args.push("-trial");
    return runSigner(args);
}

// The limits a key can set; one left out is no limit.
export type KeyLimits = Partial<Record<"nodes" | "users" | "apps" | "functions" | "previews", number>>;

// LicensePayload is what a key says: the signer's payload, field for field -
// `licensesign sign` refuses a field it does not know.
export interface LicensePayload {
    version: 1;
    licenseId: string;
    customer: { name?: string; email?: string };
    plan: PaidPlan;
    limits?: KeyLimits;
    installations: string[];
    issuedAt: string;
    expiresAt: string;
    graceDays: number;
    trial?: boolean;
}

export interface PayloadOptions {
    plan: PaidPlan;
    installation: string;
    issuedAt: Date;
    expiresAt: Date;
    graceDays?: number;
    trial?: boolean;
    limits?: KeyLimits;
    licenseId?: string;
}

// licensePayload is a key's payload with any dates - issued in the past, run
// out already - which issue cannot make.
export function licensePayload(options: PayloadOptions): LicensePayload {
    return {
        version: 1,
        licenseId: options.licenseId ?? `lic_e2e_${options.plan}`,
        customer: { name: "HivePaaS e2e" },
        plan: options.plan,
        ...(options.limits ? { limits: options.limits } : {}),
        installations: [options.installation],
        issuedAt: options.issuedAt.toISOString(),
        expiresAt: options.expiresAt.toISOString(),
        graceDays: options.graceDays ?? 0,
        ...(options.trial ? { trial: true } : {}),
    };
}

// signKey signs a payload as it is, as `licensesign sign` does.
export function signKey(payload: LicensePayload): string {
    return runSigner(["sign", "-key", edKey, "-key", mlKey, "-in", "-"], JSON.stringify(payload));
}

// runSigner answers the key the signer writes. What it did goes to its
// standard error, kept for the error should it fail.
function runSigner(args: string[], input?: string): string {
    return execFileSync(signer, args, { input, encoding: "utf8", stdio: "pipe" }).trim();
}

// restoreColony puts back the key every other test counts on, as env/up.sh
// installs it: Colony for 30 days, the same license ID and customer.
export async function restoreColony(api: APIRequestContext): Promise<void> {
    const { installationId } = await readLicense(api);
    const key = issueKey({
        plan: "colony",
        installation: installationId,
        days: 30,
        licenseId: "lic_e2e",
        customerName: "HivePaaS e2e",
        customerEmail: "admin@e2e.localhost",
    });
    const license = await installLicense(api, key);
    expect(`${license.plan} ${license.status}`, "the Colony key back in force").toBe("colony valid");
}
