import type { AppImageScan, ScannedImage, VulnSummary } from "~/projects/api/services";
import type { ImageScanApp, SystemImageScanSettings } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type SystemImageScan_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type SystemImageScan_FindOne_Res = ApiResponseBase<SystemImageScanSettings>;

export type SystemImageScan_UpdateOne_Payload = {
    updateVer: number;
    enabled: boolean;
    /** Required while scanning is on. */
    volume: { id: string } | null;
    dbRepository: string;
    javaDbRepository: string;
    schedule: {
        interval: string;
        initialTime?: Date;
    };
    notification: {
        success: { id: string };
        successUseDefault: boolean;
        failure: { id: string };
        failureUseDefault: boolean;
    };
    /** In cores; 0 for the default. */
    cpuLimit: number;
    /** As "512mb"; empty for the default. */
    memoryLimit: string;
};

export type SystemImageScan_UpdateOne_Req = ApiRequestBase<{
    payload: SystemImageScan_UpdateOne_Payload;
}>;
export type SystemImageScan_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type SystemImageScan_Execute_Req = ApiRequestBase<Record<string, never>>;
export type SystemImageScan_Execute_Res = ApiResponseBase<{
    task: {
        id: string;
    };
}>;

export type SystemImageScan_FindApps_Req = ApiRequestBase<{ pageOffset?: number; pageLimit?: number }>;
export type SystemImageScan_FindApps_Res = ApiResponseBase<ImageScanApp[]> & { total: number };

/** A part of HivePaaS: its key, which its address names, and its name. */
export interface ImageScanComponentRef {
    key: string;
    name: string;
}

/** A part of HivePaaS and what its images carry, in counts, after the global ignores. */
export interface ScannedComponent {
    component: ImageScanComponentRef;
    scannedAt: Date;
    error: string;
    summary: VulnSummary;
    images: ScannedImage[];
}

export type SystemImageScan_FindComponents_Req = ApiRequestBase<Record<string, never>>;
export type SystemImageScan_FindComponents_Res = ApiResponseBase<{ components: ScannedComponent[] }>;

export type SystemImageScan_FindComponent_Req = ApiRequestBase<{ component: string }>;
export type SystemImageScan_FindComponent_Res = ApiResponseBase<{
    component: ImageScanComponentRef;
    /** As an app's Security tab shows it. */
    scan: AppImageScan;
}>;
