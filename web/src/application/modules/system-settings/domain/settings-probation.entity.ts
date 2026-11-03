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
 * is still settling rather than that anybody is locked out.
 *
 * It is a short wait for almost everything: replacing Traefik's task takes a few
 * seconds, and a label change is live as soon as Traefik's next poll picks it up.
 * The one long case is a change that restarts HivePaaS itself - a replica or
 * worker setting sent alongside the proxy fields - where the app has to finish
 * booting before a failed request means anything.
 */
export type SettingsPendingChange = {
    changeId: string;
    appliedAt: Date;
    confirmableFrom: Date;
    deadlineAt: Date;
};
