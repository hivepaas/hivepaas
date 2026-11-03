import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

/**
 * One scaling of an app: when, from and to how many replicas, and what it was decided from - the calls or requests
 * in flight; a function's calls and those turned away; an app's requests and CPU, in percent of its limit.
 */
export interface AppAutoscaleEvent {
    time: string;
    from: number;
    to: number;
    inFlight: number;
    calls: number;
    throttled: number;
    requests: number;
    cpu: number;
    reason: string;
}

/**
 * An app's autoscale. A function scales on its calls, at `target` of its Concurrency; any other app on its requests,
 * at `requestsTarget` in flight an instance, and its CPU, at `cpuTarget` of its limit - 0 for one it does not scale on.
 */
export interface AppAutoscale {
    isFunction: boolean;
    enabled: boolean;
    minReplicas: number;
    maxReplicas: number;
    /** A function's: the share of an instance's Concurrency it keeps busy, in percent. */
    target: number;
    requestsTarget: number;
    cpuTarget: number;
    /** How long the load stays low before it scales in, such as "5m". */
    scaleInDelay: string;
    /** The app's replicas now; 0 when it is stopped. */
    replicas: number;
    /**
     * Why autoscale cannot act, on or off: what it scales on cannot be read (a reason of the logs, of the access log
     * or of the agent), `not-replicated`, or `host-ports`.
     */
    paused: string | null;
    /** Why an app's signal cannot be read now, whether it scales on it or not; null when it can. */
    requestsUnavailable: string | null;
    cpuUnavailable: string | null;
    /** Its tasks wanted and not running: the cluster may have no room for them. */
    pending: number;
    /** It writes to a volume or a bind mount its replicas would share, or each have their own. */
    writableMounts: boolean;
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
    requestsTarget: number;
    cpuTarget: number;
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
