/**
 * Where one thing a new installation still has to do stands, worked out by the
 * server from what exists.
 */
export type SetupChecklistItemStatus = "todo" | "obtaining" | "failed" | "done";

export interface SetupChecklistItem {
    status: SetupChecklistItemStatus;
    /** The dashboard's domain, for the certificate item. */
    domain: string;
    /** Why the last attempt failed, for the certificate item. */
    error: string;
}

/**
 * What the Get started card lists, for an admin while the installation step is
 * `hivepaas/get-started`.
 */
export interface SetupChecklist {
    dashboardCert: SetupChecklistItem;
    twoFactor: SetupChecklistItem;
    githubApp: SetupChecklistItem;
}
