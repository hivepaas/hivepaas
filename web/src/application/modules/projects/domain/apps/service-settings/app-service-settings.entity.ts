import type { EAppServicePlacement, EServiceMode } from "~/projects/module-shared/enums";

type ServiceModeGlobal = {
    mode: typeof EServiceMode.Global | typeof EServiceMode.GlobalJob;
};

type ServiceModeReplicated = {
    mode: typeof EServiceMode.Replicated;
    serviceReplicas: number | null;
};

type ServiceModeReplicatedJob = {
    mode: typeof EServiceMode.ReplicatedJob;
    jobMaxConcurrent: number | null;
    jobTotalCompletions: number | null;
};

export type ServiceModeSpec = ServiceModeGlobal | ServiceModeReplicated | ServiceModeReplicatedJob;

export type PlacementConstraint = {
    name: string;
    value: string;
    op: EAppServicePlacement;
};

export type PlacementPreference = {
    name: string;
    value: string;
};

export type Placement = {
    constraints: PlacementConstraint[];
    preferences: PlacementPreference[];
};

export type AppServiceSettings = {
    modeSpec: ServiceModeSpec;
    placement: Placement | null;
    updateVer: number;
};
