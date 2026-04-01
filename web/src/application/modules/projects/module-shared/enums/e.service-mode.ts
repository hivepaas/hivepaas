export const EServiceMode = {
    Replicated: "replicated",
    ReplicatedJob: "replicated-job",
    Global: "global",
    GlobalJob: "global-job",
} as const;

export type EServiceMode = (typeof EServiceMode)[keyof typeof EServiceMode];
