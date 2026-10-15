import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

/** How bad a vulnerability is, as Trivy grades it. */
export type VulnSeverity = "CRITICAL" | "HIGH" | "MEDIUM" | "LOW" | "UNKNOWN";

/** How many findings of a severity, and how many of them have a fixed version. */
export interface VulnCount {
    total: number;
    fixable: number;
}

/** Findings by severity, after ignores; `ignored` is how many were set aside. */
export interface VulnSummary {
    critical: VulnCount;
    high: VulnCount;
    medium: VulnCount;
    low: VulnCount;
    unknown: VulnCount;
    ignored: number;
}

/** An image an app runs, and its last scan. */
export interface ScannedImage {
    imageId: string;
    imageRef: string;
    platform: string;
    osFamily: string;
    osName: string;
    /** The OS is past its support: its packages get no fixes. */
    osEosl: boolean;
    /** done or failed. */
    status: string;
    /** Where the last scan read the image: "node", the node running it; "registry", its registry. */
    source: string;
    /** Why the last scan failed; the findings are the scan's before. */
    error: string;
    scannedAt: Date;
    updatedAt: Date;
    dbUpdatedAt: Date | null;
    trivyVersion: string;
    summary: VulnSummary;
}

/** One vulnerability of one package of an image. */
export interface VulnFinding {
    imageId: string;
    vulnId: string;
    pkgName: string;
    pkgType: string;
    pkgPath: string;
    installedVersion: string;
    fixedVersion: string;
    severity: VulnSeverity;
    title: string;
    url: string;
    publishedAt: Date | null;
}

/** Where an ignore was set: the app, its env, its project, or the installation ("" or "global"). */
export type VulnIgnoreScope = "app" | "project-env" | "project" | "global";

/** The ignore that sets a finding aside. */
export interface VulnIgnoreRef {
    id: string;
    scope: VulnIgnoreScope;
    /** What it is set on in its scope: an app's own and its parent app's are both of the app scope. */
    objectId: string;
    package: string;
    reason: string;
    setBy: string;
    expireAt: Date | null;
    /** The apps of the scope's children have it too; a global one that is not is HivePaaS's alone. */
    inheritable: boolean;
}

export interface IgnoredFinding extends VulnFinding {
    ignore: VulnIgnoreRef;
}

/** What an app's images carry: its Security tab. */
export interface AppImageScan {
    /** Image scanning is switched on. */
    enabled: boolean;
    /** When the app's images were last looked for; null before the first scan. */
    scannedAt: Date | null;
    /** Why none of its images could be scanned then. */
    error: string;
    summary: VulnSummary;
    images: ScannedImage[];
    /** Those no ignore sets aside, the worst first. */
    findings: VulnFinding[];
    ignored: IgnoredFinding[];
    /** The app's scan queued or running. */
    task: { id: string; status: string } | null;
}

type AppRef = { projectID: string; env: string; appID: string };

export type AppImageScan_FindOne_Req = ApiRequestBase<AppRef>;
export type AppImageScan_FindOne_Res = ApiResponseBase<AppImageScan>;

export type AppImageScan_Scan_Req = ApiRequestBase<AppRef>;
export type AppImageScan_Scan_Res = ApiResponseBase<{ task: { id: string } }>;

export interface VulnIgnorePayload {
    vulnId: string;
    package: string;
    reason: string;
    expireAt: Date | null;
    inheritable: boolean;
}

export type AppImageScan_CreateIgnore_Req = ApiRequestBase<
    AppRef & {
        scope: VulnIgnoreScope;
        payload: VulnIgnorePayload;
    }
>;
export type AppImageScan_CreateIgnore_Res = ApiResponseBase<{ id: string }>;

export type AppImageScan_DeleteIgnore_Req = ApiRequestBase<
    AppRef & {
        scope: VulnIgnoreScope;
        ignoreID: string;
    }
>;
export type AppImageScan_DeleteIgnore_Res = ApiResponseBase<{ type: "success" }>;

/** An app and what its images carry, in counts: a row of a project's Security tab. */
export interface ScannedApp {
    app: { id: string; name: string };
    project: { id: string; name: string };
    projectEnv: { id: string; key: string; name: string };
    scannedAt: Date;
    error: string;
    summary: VulnSummary;
    images: ScannedImage[];
}

export type ProjectImageScan_Find_Req = ApiRequestBase<{ projectID: string }>;
export type ProjectImageScan_Find_Res = ApiResponseBase<{
    /** Image scanning is on: the apps are scanned after each deployment and every day. */
    enabled: boolean;
    /** The project's scanned apps, the worst first. */
    apps: ScannedApp[];
}>;

export type AppImageScan_DownloadSBOM_Req = ApiRequestBase<AppRef & { imageId: string }>;
export type AppImageScan_DownloadSBOM_Res = ApiResponseBase<{
    /** The image's packages, as CycloneDX JSON. */
    blob: Blob;
    filename?: string;
}>;
