import type { License } from "@application/shared/entities";
import { ELicenseLimit } from "@application/shared/enums";
import { licenseFeatureCopy, licenseLimitCopy, licenseLimitNoun, licensePlanName } from "@application/shared/utils";

import type { PlanLimitRefusal } from "../types";

/** What the plan-limit dialog says about a refusal. */
export interface PlanLimitReachedCopy {
    title: string;
    body: string;
}

/**
 * The dialog's title and text for a refusal: what the plan allows of the kind,
 * what counts, and what makes room.
 *
 * A kind this dashboard does not know, from a newer HivePaaS, and a refusal
 * whose params could not be read are told in the server's own words: its
 * message names the kind, in the person's language.
 */
export function planLimitReachedCopy(refusal: PlanLimitRefusal): PlanLimitReachedCopy {
    const { limit, max, plan, feature } = refusal;

    if (feature) {
        const copy = licenseFeatureCopy(feature);
        return {
            title: `${capitalize(copy.name)} need a paid plan`,
            body: `The ${licensePlanName(plan)} plan does not include ${copy.name}. They come with ${copy.plans}.`,
        };
    }

    if (max === null) {
        return serverCopy(refusal);
    }

    const planName = licensePlanName(plan);

    // Not a count: Bee allows none to be added, and every paid plan allows them all.
    if (limit === ELicenseLimit.SignInProviders) {
        return {
            title: "Google, Microsoft and OpenID Connect sign-in need a paid plan",
            body: `${planName} does not include them. The ones already set up keep signing people in; none can be added or enabled.`,
        };
    }

    const copy = countedCopy(limit, max, planName);

    if (copy === null) {
        return serverCopy(refusal);
    }

    // A kind the plan leaves out, where removing one would make no room.
    if (max === 0) {
        return { title: copy.title, body: `The ${planName} plan does not include ${licenseLimitCopy(limit).many}.` };
    }

    return copy;
}

/** The copy for a kind the plan counts, or null for one this dashboard does not know. */
function countedCopy(limit: string, max: number, planName: string): PlanLimitReachedCopy | null {
    const allowed = `${max} ${licenseLimitNoun(limit, max)}`;

    switch (limit) {
        case ELicenseLimit.Nodes:
            return {
                title: `No more nodes on ${planName}`,
                body: `The ${planName} plan allows ${allowed}, and the swarm has as many, so HivePaaS gives no join command. Removing a node from the swarm frees its place; draining it does not.`,
            };
        case ELicenseLimit.Users:
            return {
                title: `No more users on ${planName}`,
                body: `The ${planName} plan allows ${allowed}, counting those invited who have not signed up yet. Disable a user or withdraw an invitation to free a seat.`,
            };
        case ELicenseLimit.Apps:
            return {
                title: `No more apps on ${planName}`,
                body: `The ${planName} plan allows ${allowed}. Previews, functions, the databases a template or a preview makes, and HivePaaS's own apps are not counted. Delete an app to make room.`,
            };
        case ELicenseLimit.Functions:
            return {
                title: `No more functions on ${planName}`,
                body: `The ${planName} plan allows ${allowed}. Delete one to make room.`,
            };
        case ELicenseLimit.Previews:
            return {
                title: `No more previews at a time on ${planName}`,
                body: `The ${planName} plan runs ${allowed} at a time. Close one to open another.`,
            };
        default:
            return null;
    }
}

function capitalize(text: string): string {
    return text.charAt(0).toUpperCase() + text.slice(1);
}

function serverCopy({ message }: PlanLimitRefusal): PlanLimitReachedCopy {
    return {
        title: "Plan limit reached",
        body: message !== "" ? message : "This installation's plan does not allow it.",
    };
}

/**
 * How many of the refused kind there are now, against the most the plan in force
 * allows: "Now 7 of 5 nodes on Comb". It is read from the license fetched as the
 * dialog opened, so it counts what is there now.
 *
 * Undefined when there is nothing to count against: no license - it is loading,
 * it failed, or this person cannot read System - a kind the plan no longer
 * limits or leaves out, and sign-in providers, which are not a count.
 */
export function planLimitUsageLine(refusal: PlanLimitRefusal, license: License | undefined): string | undefined {
    if (license === undefined || refusal.limit === ELicenseLimit.SignInProviders) {
        return undefined;
    }

    const usage = license.limits.find(item => item.limit === refusal.limit);

    // A max of null or 0: not limited, or left out.
    if (!usage?.max) {
        return undefined;
    }

    const noun = licenseLimitNoun(usage.limit, usage.max);

    return `Now ${usage.used} of ${usage.max} ${noun} on ${licensePlanName(license.plan)}`;
}
