/**
 * Something across the cluster that needs attention, already narrowed by the
 * server to what the signed-in user may see.
 */
export const AttentionKind = {
    AppNotRunning: "app-not-running",
    AppRestarting: "app-restarting",
    AppVulnerable: "app-vulnerable",
    NodeDown: "node-down",
    NodeOvercommitted: "node-overcommitted",
} as const;

export type AttentionKind = (typeof AttentionKind)[keyof typeof AttentionKind];

export const AttentionScope = {
    App: "app",
    System: "system",
    Cluster: "cluster",
} as const;

export type AttentionScope = (typeof AttentionScope)[keyof typeof AttentionScope];

export interface AttentionItem {
    /** Kept as a string: a kind this screen does not know yet is still shown, plainly. */
    kind: string;
    severity: "critical" | "warning";
    scope: string;
    project: { id: string; name: string } | null;
    env: string;
    app: { id: string; name: string } | null;
    /** What it is about: an app, a system service, a node. */
    subject: string;
    running: number;
    desired: number;
    restarts: number;
    lastError: string;
    nodeState: string;
    memoryLimitsBytes: number;
    memoryTotalBytes: number;
    /** Of an app vulnerable: how many fixable CRITICAL vulnerabilities its images carry. */
    criticalFixable: number;
    since: Date | null;
    /** Whether the user may change it on the screen it leads to, beyond looking. */
    canAct: boolean;
}
