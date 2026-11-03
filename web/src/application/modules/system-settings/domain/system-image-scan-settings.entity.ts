import type { SettingsBaseEntity } from "~/settings/domain";

import type { SystemSslRenewalNotification } from "./system-ssl-renewal-settings.entity";

/**
 * The volume Trivy keeps its data on, and so where scans run: on the node a
 * pinned one is on, on a node carrying its label, or - shared - on the node
 * running each image.
 */
export interface ImageScanVolume {
    id: string;
    name: string;
    /** The volume is gone: no scan can run. */
    missing: boolean;
    nodeId: string;
    nodeName: string;
    nodeLabel: string;
    shared: boolean;
    /** Why no scan can run on it now: no ready node carries its label. */
    problem: string;
}

/** Where scans run, and what the last one used. */
export interface ImageScannerStatus {
    volume: ImageScanVolume | null;
    image: string;
    trivyVersion: string;
    dbUpdatedAt: Date | null;
    lastScanAt: Date | null;
    /** The last scan's error, when it failed. */
    lastError: string;
}

/** Image scanning's settings: the addon and its daily scan. */
export interface SystemImageScanSettings extends SettingsBaseEntity {
    enabled: boolean;
    /** The cluster volume Trivy keeps its DBs and its layer cache on. */
    volume: { id: string } | null;
    dbRepository: string;
    javaDbRepository: string;
    schedule: { interval: string; initialTime: Date | null };
    notification: SystemSslRenewalNotification | null;
    /** What each scan's container runs under, in cores and as "512mb". */
    cpuLimit: number;
    memoryLimit: string;
    nextRuns: Date[];
    scannerStatus: ImageScannerStatus | null;
}

/** An app of the apps list: what its images carry, in counts. */
export interface ImageScanApp {
    app: { id: string; name: string };
    project: { id: string; name: string };
    /** key is what an app's address names. */
    projectEnv: { id: string; key: string; name: string };
    scannedAt: Date;
    error: string;
    summary: {
        critical: { total: number; fixable: number };
        high: { total: number; fixable: number };
        medium: { total: number; fixable: number };
        low: { total: number; fixable: number };
        unknown: { total: number; fixable: number };
        ignored: number;
    };
    images: {
        imageId: string;
        imageRef: string;
        osFamily: string;
        osName: string;
        osEosl: boolean;
        status: string;
        error: string;
        scannedAt: Date;
    }[];
}
