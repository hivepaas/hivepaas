/**
 * A spec is a snapshot of configuration - settings, plus the parts of an app
 * that live in its Swarm service - packed into one tar.gz of YAML.
 */

export type SpecExportScope =
    | { type: "global" }
    | { type: "project"; projectID: string }
    | { type: "project-env"; projectID: string; projectEnvID: string }
    | { type: "app"; projectID: string; projectEnvID: string; appID: string };

/**
 * What happens to the values settings keep encrypted.
 *
 * There is deliberately no mode that leaves the stored ciphertext in place: it
 * is readable only by the installation holding that data key, so it would look
 * like a secret and silently fail anywhere else.
 */
export type SpecSecretsMode = "omit" | "encrypted" | "plaintext";

/**
 * The counts the server returns in the X-HivePaaS-Spec-Report header.
 *
 * The detail is inside the bundle as report.yaml rather than here - the full
 * report already exceeded what a header can safely carry on a small
 * installation.
 */
export interface SpecExportSummary {
    files: number;
    issues: number;
    bySeverity?: Record<string, number>;
    byCode?: Record<string, number>;
    reportFile?: string;
}

export interface SpecExportResult {
    filename: string;
    sizeBytes: number;
    summary?: SpecExportSummary;
}
