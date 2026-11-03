import { ROUTE } from "@application/shared/constants";
import type { License, LicenseLimitUsage } from "@application/shared/entities";
import { ELicenseLimit, ELicensePlan, ELicenseStatus } from "@application/shared/enums";
import {
    daysLeft,
    daysLeftText,
    formatLicenseDate,
    licenseLimitCopy,
    licenseLimitNoun,
    licensePlanName,
    overLimits,
} from "@application/shared/utils";

/** How much a banner asks of the reader: a heads-up, something to see to soon, or something already lost. */
export type LicenseBannerTone = "info" | "warning" | "danger";

export interface LicenseBannerLink {
    label: string;
    /** A page of the license portal when external, otherwise a dashboard route. */
    href: string;
    /** On the license portal: it opens in a new tab. */
    external: boolean;
}

export interface LicenseBannerModel {
    /**
     * The situation it is about, not only its kind: a dismissal hides this id
     * alone, so a new situation - another key, another date, another limit gone
     * over - shows again however recently the last one was dismissed.
     */
    id: string;
    /**
     * A status banner says what the license's status means, which System ›
     * License says too, so it is left out there. The limits banner says what is
     * over the plan's limits.
     */
    kind: "status" | "limits";
    tone: LicenseBannerTone;
    title: string;
    body: string;
    links: LicenseBannerLink[];
    /** How long dismissing it hides it, in this browser. */
    snoozeMs: number;
}

const HOUR_MS = 3_600_000;
const DAY_MS = 24 * HOUR_MS;

/*
 * A banner counting down to a date - a trial, a grace period, a moved key still
 * holding - comes back the next day: by then there is a day less. One about what
 * has already happened - the installation is on Bee, or over a limit - comes
 * back a week later: it does not change by the day, and a banner back every
 * morning is one people learn to look past.
 */
const SNOOZE_COUNTDOWN_MS = DAY_MS;
const SNOOZE_SETTLED_MS = 7 * DAY_MS;

const ON_BEE_BODY = "Everything keeps running; nothing more can be added past Bee's limits.";

/**
 * The banners the license calls for at now, the status banner first: at most
 * two. Pure, so that it can be tested: now is passed in, and dismissals and the
 * page shown are left to the caller.
 *
 * Days are counted between instants, so the browser's time zone does not change
 * them. The status is the server's, and authoritative: a browser whose clock is
 * ahead can only make a count read "ends now" early. The server always sends the
 * date a status counts down to; were it missing, there would be no time to count.
 */
export function licenseBannersOf(license: License, now: number): LicenseBannerModel[] {
    return [statusBannerOf(license, now), limitsBannerOf(license)].filter(
        (banner): banner is LicenseBannerModel => banner !== null,
    );
}

/** How long a dismissal lasts, as its button's hint says it: "24 hours", "7 days". */
export function licenseBannerSnoozeText(snoozeMs: number): string {
    return snoozeMs < 2 * DAY_MS ? `${Math.round(snoozeMs / HOUR_MS)} hours` : `${Math.round(snoozeMs / DAY_MS)} days`;
}

function statusBannerOf(license: License, now: number): LicenseBannerModel | null {
    const { expiresAt, graceEndsAt, licenseId, movedAt } = license;

    switch (license.status) {
        case ELicenseStatus.Trial:
            return {
                id: `trial:${licenseId}:${stamp(expiresAt)}`,
                kind: "status",
                tone: "info",
                title: trialTitle(license, now),
                body: "Then this installation is on Bee, the free plan.",
                links: [...upgradeLinks(license), licenseLink()],
                snoozeMs: SNOOZE_COUNTDOWN_MS,
            };

        case ELicenseStatus.Grace:
            return {
                id: `grace:${licenseId}:${stamp(expiresAt)}`,
                kind: "status",
                tone: "warning",
                title: graceTitle(license, now),
                body: "Renew it, or paste the renewed key.",
                links: [...portalLinks("Renew", license.portal?.accountUrl), licenseLink()],
                snoozeMs: SNOOZE_COUNTDOWN_MS,
            };

        case ELicenseStatus.Expired:
            return {
                id: `expired:${licenseId}:${stamp(expiresAt)}`,
                kind: "status",
                tone: "danger",
                title: "The license has expired: this installation is on Bee.",
                body: ON_BEE_BODY,
                links: [...upgradeLinks(license), licenseLink()],
                snoozeMs: SNOOZE_SETTLED_MS,
            };

        case ELicenseStatus.Moved:
            // A moved key holds for its grace, counted from the move, and then
            // the plan in force turns to Bee while the status stays moved: it
            // never becomes expired. No key names Bee, so the plan says which.
            if (license.plan === ELicensePlan.Bee) {
                return {
                    id: `moved-bee:${licenseId}:${stamp(movedAt)}`,
                    kind: "status",
                    tone: "danger",
                    title: "This license key now belongs to another installation: this installation is on Bee.",
                    body: ON_BEE_BODY,
                    links: [licenseLink()],
                    snoozeMs: SNOOZE_SETTLED_MS,
                };
            }

            return {
                id: `moved:${licenseId}:${stamp(movedAt)}`,
                kind: "status",
                tone: "warning",
                title: "This license key now belongs to another installation.",
                body: movedHoldText(graceEndsAt, now),
                links: [licenseLink()],
                snoozeMs: SNOOZE_COUNTDOWN_MS,
            };

        case ELicenseStatus.OtherInstallation:
            // This installation's ID is in the id: a swarm made again, as in a
            // disaster recovery, has a new one, and is a new situation.
            return {
                id: `other-installation:${licenseId}:${license.installationId}`,
                kind: "status",
                tone: "danger",
                title: "The stored license key is for another installation: this installation is on Bee.",
                body: ON_BEE_BODY,
                links: [licenseLink()],
                snoozeMs: SNOOZE_SETTLED_MS,
            };

        case ELicenseStatus.Invalid:
            // A key that did not verify has no license ID; why it did not is
            // what tells one such key from another.
            return {
                id: `invalid:${license.reason}`,
                kind: "status",
                tone: "danger",
                title: "The stored license key cannot be verified: this installation is on Bee.",
                body: ON_BEE_BODY,
                links: [licenseLink()],
                snoozeMs: SNOOZE_SETTLED_MS,
            };

        default:
            // No key, a key in force with time to run - there is no "expiring
            // soon" banner: the renewed key arrives by itself, and System ›
            // License shows the last check's error - or a status a newer
            // HivePaaS added, which this dashboard cannot put into words.
            return null;
    }
}

/**
 * A single banner for all the limits gone over, as after a key ran out. Sign-in
 * providers are never among them: see overLimits.
 *
 * Its id names each limit with its max, and not with the count: one more app
 * over the same limit is the same situation, while another limit gone over, or
 * another plan's max, is a new one.
 */
function limitsBannerOf(license: License): LicenseBannerModel | null {
    const over = overLimits(license);

    if (over.length === 0) {
        return null;
    }

    return {
        id: `limits:${over.map(usage => `${usage.limit}=${usage.max ?? ""}`).join(",")}`,
        kind: "limits",
        tone: "warning",
        title: `${over.map(overLimitClause).join("; ")}.`,
        body: "Nothing is removed. Remove some, or upgrade.",
        links: [...upgradeLinks(license), licenseLink()],
        snoozeMs: SNOOZE_SETTLED_MS,
    };
}

/** "7 of 5 nodes: no node can be added". The noun goes with the max: "2 of 1 app". */
function overLimitClause(usage: LicenseLimitUsage): string {
    // overLimits leaves out the kinds with no max.
    const max = usage.max ?? 0;

    return `${usage.used} of ${max} ${licenseLimitNoun(usage.limit, max)}: ${overLimitConsequence(usage.limit)}`;
}

/** What can no longer be done while there are more of a kind than the plan allows. */
function overLimitConsequence(limit: string): string {
    switch (limit) {
        case ELicenseLimit.Users:
            return "no user can be invited or re-enabled";
        case ELicenseLimit.Previews:
            return "no preview can be opened";
        default:
            return `no ${licenseLimitCopy(limit).one} can be added`;
    }
}

/** "Hive trial: 3 days left." The plan is the key's, which is the plan in force. */
function trialTitle(license: License, now: number): string {
    const left = license.expiresAt ? daysLeftText(license.expiresAt, now) : "ends now";

    return `${licensePlanName(license.keyPlan || license.plan)} trial: ${left}.`;
}

/** "The Comb license ran out on Oct 3, 2026: 13 days of grace left." */
function graceTitle(license: License, now: number): string {
    const on = license.expiresAt ? ` on ${formatLicenseDate(license.expiresAt)}` : "";

    return `The ${licensePlanName(license.plan)} license ran out${on}: ${graceLeftText(license.graceEndsAt, now)}.`;
}

/** "13 days of grace left", "1 day of grace left", or "its grace ends now" once the moment has come. */
function graceLeftText(graceEndsAt: Date | null, now: number): string {
    const days = graceEndsAt ? daysLeft(graceEndsAt, now) : 0;

    if (days === 0) {
        return "its grace ends now";
    }

    return days === 1 ? "1 day of grace left" : `${days} days of grace left`;
}

/** How long a moved key still holds here. */
function movedHoldText(graceEndsAt: Date | null, now: number): string {
    const days = graceEndsAt ? daysLeft(graceEndsAt, now) : 0;

    if (days === 0) {
        return "It stops holding here now; then this installation is on Bee.";
    }

    return `It holds here ${days} more ${days === 1 ? "day" : "days"}, then this installation is on Bee.`;
}

function licenseLink(): LicenseBannerLink {
    return { label: "License", href: ROUTE.systemSettings.license.$route, external: false };
}

function upgradeLinks(license: License): LicenseBannerLink[] {
    return portalLinks("Upgrade", license.portal?.pricingUrl);
}

/** A link to a page of the portal, when the server named it. */
function portalLinks(label: string, href: string | undefined): LicenseBannerLink[] {
    return href ? [{ label, href, external: true }] : [];
}

/** A date as it goes into a banner's id. */
function stamp(date: Date | null): string {
    return date?.toISOString() ?? "";
}
