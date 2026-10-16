import type { ELicenseLimit, ELicensePlan, ELicenseRefreshResult, ELicenseStatus } from "@application/shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

/**
 * The license in force on this installation, with how much of each limited kind
 * there is: what System › License shows, and what the banners, the plan-limit
 * dialog and the OAuth page read.
 *
 * The key's own fields - keyPlan, trial, licenseId, customerName, installations,
 * issuedAt and expiresAt - are filled whenever the key's signatures verified, also
 * when it is not in force, so the page can say which key it is. The enum-like
 * fields stay strings, so a value from a newer HivePaaS still arrives.
 */
export interface License {
    /** What the stored key amounts to here, now. */
    status: OpenApiConstant<ELicenseStatus>;
    /** The plan in force: the key's, or Bee when the key puts nothing in force. */
    plan: OpenApiConstant<ELicensePlan>;
    /** The plan the key was issued for, in force or not. Empty when there is no key, or it did not verify. */
    keyPlan: OpenApiConstant<ELicensePlan>;
    trial: boolean;
    /** Signed by the development keys: only a HivePaaS in development accepts it. */
    development: boolean;
    licenseId: string;
    /** Of the customer only the name is sent; the email a key carries never leaves the server. */
    customerName: string;
    /** The ID a key names this installation by: its swarm's. Empty when Docker could not say. */
    installationId: string;
    /** The installations the key is for. */
    installations: string[];
    issuedAt: Date | null;
    expiresAt: Date | null;
    /** When a key for this installation stops holding: its grace after it expires, or after it was moved if that comes first. */
    graceEndsAt: Date | null;
    /** When the portal moved the key to another installation. */
    movedAt: Date | null;
    /** Why the key is invalid. */
    reason: string;
    /** Every kind a plan can limit, in the same order every time. */
    limits: LicenseLimitUsage[];
    portal: LicensePortal | null;
    refresh: LicenseRefreshInfo | null;
}

/** One kind a plan limits: the most the plan in force allows, and how many there are. */
export interface LicenseLimitUsage {
    limit: OpenApiConstant<ELicenseLimit>;
    /**
     * Null when the plan in force does not limit the kind. Sign-in providers are 0
     * (Bee: none can be added) or null (every one can).
     */
    max: number | null;
    /** Counted as a creation counts them. For sign-in providers, the paid ones that are active. */
    used: number;
}

/** The license portal's pages. Those for buying a key and starting a trial carry this installation's ID, when it is known. */
export interface LicensePortal {
    url: string;
    pricingUrl: string;
    trialUrl: string;
    /** Where the keys bought are kept. */
    accountUrl: string;
}

/** The daily check with the portal for a newer key, and how the last check went. */
export interface LicenseRefreshInfo {
    /** The daily check is on. When it is off, a person can still ask now. */
    enabled: boolean;
    /** When the portal was last asked, whatever came of it. */
    lastCheckedAt: Date | null;
    /** Empty until the first check. */
    lastResult: OpenApiConstant<ELicenseRefreshResult>;
    /** Why the last check failed. */
    lastError: string;
}
