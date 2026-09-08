/**
 * A settings change that has been applied but not yet vouched for.
 *
 * HivePaaS applies the changes that can lock you out on trial: unless a request
 * gets back in through the new configuration and confirms it, the change is
 * undone at `deadlineAt`. That is the guard for the mistakes nothing can detect
 * up front - a broken auth reference, a domain whose DNS is not ready, an
 * allowlist or a proxy depth that excludes the person editing it.
 *
 * `confirmableFrom` is the server saying when the change should be live. Before
 * it, a request would have travelled through the configuration being replaced, so
 * confirming proves nothing - and being unable to reach HivePaaS at all means it
 * is still settling rather than that anybody is locked out. The gap is short for
 * routing changes, which only rewrite labels, and much longer for service
 * settings, which restart traefik.
 */
export type SettingsPendingChange = {
    changeId: string;
    appliedAt: Date;
    confirmableFrom: Date;
    deadlineAt: Date;
};
