import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

/** One scaling of a function: when, from and to how many replicas, and what it was decided from. */
export interface AppAutoscaleEvent {
    time: string;
    from: number;
    to: number;
    inFlight: number;
    calls: number;
    throttled: number;
    reason: string;
}

/** Why autoscale cannot act: its calls cannot be read, or the function runs no set number of instances. */
export type AppAutoscalePausedReason =
    | "disabled"
    | "apps-not-collected"
    | "no-query-endpoint"
    | "driver-unreadable"
    | "identity-missing"
    | "not-replicated";

export interface AppAutoscale {
    enabled: boolean;
    minReplicas: number;
    maxReplicas: number;
    /** The share of an instance's Concurrency it keeps busy, in percent. */
    target: number;
    /** How long the load stays low before it scales in, such as "5m". */
    scaleInDelay: string;
    /** The function's replicas now; 0 when it is stopped. */
    replicas: number;
    paused: AppAutoscalePausedReason | null;
    /** The latest scalings, the latest first. */
    events: AppAutoscaleEvent[];
    updateVer: number;
}

export type AppAutoscale_FindOne_Req = ApiRequestBase<{ projectID: string; env: string; appID: string }>;
export type AppAutoscale_FindOne_Res = ApiResponseBase<AppAutoscale>;

export type AppAutoscale_UpdatePayload = {
    enabled: boolean;
    minReplicas: number;
    maxReplicas: number;
    target: number;
    scaleInDelay: string;
    updateVer: number;
};

export type AppAutoscale_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    payload: AppAutoscale_UpdatePayload;
}>;
export type AppAutoscale_UpdateOne_Res = ApiResponseBase<{ type: "success"; warning: string | null }>;
