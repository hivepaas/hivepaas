import type { SpecExportScope, SpecSecretsMode } from "./spec-export.entity";

/**
 * Importing a spec: the bundle export produced, planned against this
 * installation by validate, then applied.
 *
 * An import writes at the scope of the screen it is started from - the whole
 * installation, a project, or one env of it. A wider bundle is fine: the part
 * inside the scope is taken.
 */
export type SpecImportScope = Exclude<SpecExportScope, { type: "app" }>;

/** What import does with an object this installation already has. */
export type SpecImportExisting = "update" | "keep";

export interface SpecImportOptions {
    existing: SpecImportExisting;
    /** Queue the first deployment of every app the import creates. */
    deployCreated: boolean;
    /** Queue a deployment for every updated app whose deployment source changed. */
    deployChangedSource: boolean;
}

/**
 * The part of a bundle an import takes, as node paths. A path takes its node and
 * everything below it; an empty include takes everything.
 *
 * A node the include leaves out is still one the import may pull in, when
 * something it takes refers to it and this installation lacks it. A node in
 * exclude is never pulled in.
 */
export interface SpecImportSelection {
    include: string[];
    exclude: string[];
}

export type SpecImportNodeKind = "global" | "project" | "env" | "settings" | "app";

export type SpecImportNodeAction = "create" | "update" | "unchanged" | "keep" | "skip";

export type SpecImportNodeOutcome = "applied" | "unchanged" | "skipped" | "failed";

/**
 * blocked stops the import; skipped leaves one object out; fixable clears part
 * of one; warning changes nothing but may not be what was meant. All but blocked
 * need accepting.
 */
export type SpecImportSeverity = "blocked" | "skipped" | "fixable" | "warning";

export interface SpecImportIssue {
    /** Absent for a note, which says what import does and needs no accepting. */
    severity?: SpecImportSeverity;
    code: string;
    path: string;
    detail?: Record<string, unknown>;
    /** The node of the bundle holding what the issue could not reach. */
    availableIn?: string;
    /** What import does about it, in a sentence. */
    action?: string;
    hint?: string;
}

export interface SpecImportNode {
    path: string;
    kind: SpecImportNodeKind;
    key?: string;
    name?: string;
    selected: boolean;
    /** "user", or "dependency" when the import pulled it in. */
    selectedBy?: string;
    action: SpecImportNodeAction;
    matchedBy?: string;
    /** What differs, by name - never by value. */
    changes: string[];
    restart: boolean;
    deploy: boolean;
    issues: SpecImportIssue[];
    notes: SpecImportIssue[];
    /** What apply did; absent in validate's plan. */
    outcome?: SpecImportNodeOutcome;
    /** Why an app failed after the import was saved. */
    error?: string;
}

export interface SpecImportBundleInfo {
    apiVersion: string;
    scope: string;
    exportedAt: Date;
    sourceAppVersion: string;
    secretsMode: SpecSecretsMode;
    digest: string;
}

export interface SpecImportPlan {
    bundle: SpecImportBundleInfo;
    nodes: SpecImportNode[];
    summary: Record<string, number>;
    /** Binds apply to this plan: a plan that changed since is refused. */
    planHash: string;
}

export interface SpecImportDeployment {
    appId: string;
    deploymentId: string;
}

export interface SpecImportResult {
    plan: SpecImportPlan;
    deployments: SpecImportDeployment[];
    /** What failed after the import was saved: it stands, and this says what did not follow. */
    warning?: string;
}
