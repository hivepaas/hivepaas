import type { License } from "@application/shared/entities";
import { ELicensePlan, ELicenseRefreshResult, ELicenseStatus } from "@application/shared/enums";
import { formatLicenseDate, licensePlanName } from "@application/shared/utils";

/** Whether the installation is on Bee now: no key, or a key that puts nothing in force. */
export function isOnBee(license: License): boolean {
    return license.plan === ELicensePlan.Bee;
}

const HOLDING_STATUSES: ReadonlySet<string> = new Set<string>([
    ELicenseStatus.Valid,
    ELicenseStatus.Trial,
    ELicenseStatus.Grace,
    ELicenseStatus.Moved,
]);

/**
 * Whether a key is in force, which a key installed now replaces whatever its
 * plan: one within its period or its grace, or a moved one until the grace the
 * move leaves runs out. After that a moved key puts nothing in force - the plan
 * is Bee - and there is nothing to lose by replacing it.
 */
export function keyInForce(license: License): boolean {
    return HOLDING_STATUSES.has(license.status) && !isOnBee(license);
}

/**
 * What a key put in force amounts to, for the toast that says so: "Hive trial,
 * until Oct 31, 2026".
 */
export function licenseInForceSummary(license: License): string {
    const plan = `${licensePlanName(license.plan)}${license.trial ? " trial" : ""}`;

    if (license.status === ELicenseStatus.Grace && license.graceEndsAt) {
        return `${plan}, in its grace period until ${formatLicenseDate(license.graceEndsAt)}`;
    }

    return license.expiresAt ? `${plan}, until ${formatLicenseDate(license.expiresAt)}` : plan;
}

/**
 * What a check with the portal came to, as the page and its toasts say it. A
 * result from a newer HivePaaS is shown as it is.
 */
const REFRESH_RESULT_TEXT: Readonly<Record<ELicenseRefreshResult, string>> = {
    [ELicenseRefreshResult.Newer]: "A newer key was installed",
    [ELicenseRefreshResult.Moved]: "The portal moved this key to another installation",
    [ELicenseRefreshResult.Current]: "This key is the newest one",
    [ELicenseRefreshResult.Unknown]: "The portal did not issue this key, so it has none newer",
    [ELicenseRefreshResult.None]: "No key to ask about",
    [ELicenseRefreshResult.Disabled]: "The daily check is off",
    [ELicenseRefreshResult.Failed]: "The check failed",
};

const REFRESH_TEXT_BY_RESULT: ReadonlyMap<string, string> = new Map(Object.entries(REFRESH_RESULT_TEXT));

export function refreshResultText(result: string): string {
    return REFRESH_TEXT_BY_RESULT.get(result) ?? result;
}
