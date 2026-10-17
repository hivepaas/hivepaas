import type { License, LicenseLimitUsage } from "@application/shared/entities";
import { ELicenseLimit, ELicensePlan, EOAuthKind } from "@application/shared/enums";

/**
 * A plan's name as a person reads it. A plan this dashboard does not know, such
 * as one a newer HivePaaS added, is shown as it is, as the backend's
 * LicensePlan.DisplayName does.
 */
export function licensePlanName(plan: string): string {
    switch (plan) {
        case ELicensePlan.Bee:
            return "Bee";
        case ELicensePlan.Comb:
            return "Comb";
        case ELicensePlan.Hive:
            return "Hive";
        case ELicensePlan.Colony:
            return "Colony";
        case ELicensePlan.Apiary:
            return "Apiary";
        default:
            return plan;
    }
}

/** How the dashboard speaks of a limited kind. */
export interface LicenseLimitCopy {
    /** What its usage meter is labelled. */
    label: string;
    /** One of it: "1 node". */
    one: string;
    /** Any other number of it: "0 nodes", "5 nodes". */
    many: string;
    /** What counts towards the limit, said under the meter. Empty when there is nothing to add. */
    counts: string;
}

export const LICENSE_LIMIT_COPY: Readonly<Record<ELicenseLimit, LicenseLimitCopy>> = {
    [ELicenseLimit.Nodes]: {
        label: "Nodes",
        one: "node",
        many: "nodes",
        counts: "Every node of the swarm, whatever its state. Removing a node frees its place; draining it does not.",
    },
    [ELicenseLimit.Users]: {
        label: "Users",
        one: "user",
        many: "users",
        counts: "Active or invited. Disabled users don't count.",
    },
    [ELicenseLimit.Apps]: {
        label: "Apps",
        one: "app",
        many: "apps",
        counts: "Not counting previews, functions, an app's own databases, or HivePaaS's apps.",
    },
    [ELicenseLimit.Functions]: {
        label: "Functions",
        one: "function",
        many: "functions",
        counts: "",
    },
    [ELicenseLimit.Previews]: {
        label: "Previews at a time",
        one: "preview",
        many: "previews",
        counts: "",
    },
    // Not a count, so it has no meter and its row is worded on its own. The nouns
    // are what the server's own refusal calls these providers.
    [ELicenseLimit.SignInProviders]: {
        label: "Google, Microsoft and OpenID Connect sign-in",
        one: "Google, Microsoft or OpenID Connect sign-in provider",
        many: "Google, Microsoft or OpenID Connect sign-in providers",
        counts: "",
    },
};

const COPY_BY_LIMIT: ReadonlyMap<string, LicenseLimitCopy> = new Map(Object.entries(LICENSE_LIMIT_COPY));

/**
 * The copy for a limit as the server names it. A kind this dashboard does not
 * know, from a newer HivePaaS, is called by its name.
 */
export function licenseLimitCopy(limit: string): LicenseLimitCopy {
    return COPY_BY_LIMIT.get(limit) ?? { label: limit, one: limit, many: limit, counts: "" };
}

/** The noun for count of a limited kind: licenseLimitNoun("apps", 1) is "app", and any other count "apps". */
export function licenseLimitNoun(limit: string, count: number): string {
    const copy = licenseLimitCopy(limit);

    return count === 1 ? copy.one : copy.many;
}

const DAY_MS = 86_400_000;

/**
 * Whole days left until a moment, rounded up, and never below 0: 12 days and 23
 * hours left is 13. It counts between instants, so the browser's time zone does
 * not change it.
 */
export function daysLeft(until: Date, now: number = Date.now()): number {
    return Math.max(0, Math.ceil((until.getTime() - now) / DAY_MS));
}

/** "3 days left", "1 day left", or "ends now" once the moment has come. */
export function daysLeftText(until: Date, now: number = Date.now()): string {
    const days = daysLeft(until, now);

    if (days === 0) {
        return "ends now";
    }

    return days === 1 ? "1 day left" : `${days} days left`;
}

/**
 * The kinds there are more of than the plan in force allows, as after a key ran
 * out: nothing is removed, but none more can be added.
 *
 * Sign-in providers are left out. Bee allows none to be added while those set up
 * keep working, so "1 of 0" would only mislead; the OAuth page marks them instead.
 */
export function overLimits(license: License): LicenseLimitUsage[] {
    return license.limits.filter(
        usage => usage.limit !== ELicenseLimit.SignInProviders && usage.max !== null && usage.used > usage.max,
    );
}

/**
 * Whether a Google, Microsoft or OpenID Connect sign-in provider can be added or
 * enabled: false on Bee, whose limit for them is 0 - the backend's own rule, not
 * a check of the plan's name - and true on every paid plan.
 *
 * Undefined when that is not known: no license, as for someone who cannot read
 * System, or while it loads or after it failed. Mark nothing then; the server's
 * refusal opens the plan-limit dialog, which explains it.
 */
export function paidSignInAllowed(license?: License): boolean | undefined {
    const usage = license?.limits.find(item => item.limit === ELicenseLimit.SignInProviders);

    if (!usage) {
        return undefined;
    }

    return usage.max === null || usage.max > 0;
}

const PAID_SIGN_IN_KINDS: ReadonlySet<string> = new Set<string>([
    EOAuthKind.Google,
    EOAuthKind.MicrosoftOnline,
    EOAuthKind.OpenIDConnect,
]);

/**
 * Whether a sign-in provider's kind is one Bee leaves out: signing in with a
 * company's identity provider rather than with a code host. It mirrors
 * IsPaidSignIn in the backend's hivepaas_app/pkg/license/limits.go, and has to
 * name the same kinds.
 */
export function isPaidSignInKind(kind: string | undefined): boolean {
    return kind !== undefined && PAID_SIGN_IN_KINDS.has(kind);
}

const DATE_FORMAT: Intl.DateTimeFormatOptions = { year: "numeric", month: "short", day: "numeric" };

/**
 * A license date as the dashboard shows it: the day, in the browser's language,
 * such as "Oct 17, 2026". Put formatLicenseDateTime in its title, as a key runs
 * out at an hour rather than on a day.
 */
export function formatLicenseDate(date: Date): string {
    return date.toLocaleDateString(undefined, DATE_FORMAT);
}

/** The full date and time, for the title of a date shown with formatLicenseDate. */
export function formatLicenseDateTime(date: Date): string {
    return date.toLocaleString();
}
