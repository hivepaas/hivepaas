/**
 * Where the dashboard's certificate stands, worked out by the server from what
 * exists: `todo` while it has the installation's self-signed one.
 */
export type DashboardCertStatus = "todo" | "obtaining" | "failed" | "done";

export interface DashboardCert {
    status: DashboardCertStatus;
    /** The dashboard's domain. */
    domain: string;
    /** Why the last attempt failed. */
    error: string;
}
