export type AppResourceSettings = {
    reservations: ResourceReservations | null;
    limits: ResourceLimits | null;
    ulimits: Ulimit[];
    capabilities: Capabilities | null;
    updateVer: number;
};

export type ResourceReservations = {
    cpus?: number;
    memoryMB?: number;
    genericResources?: GenericResource[];
};

export type GenericResource = {
    kind: string;
    value: string;
};

export type ResourceLimits = {
    cpus?: number;
    memoryMB?: number;
    pids?: number;
};

export type Ulimit = {
    name: string;
    hard: number;
    soft: number;
};

export type Capabilities = {
    capabilityAdd?: string[];
    capabilityDrop?: string[];
    enableGPU?: boolean;
    oomScoreAdj?: number;
    sysctls?: Record<string, string>;
};
