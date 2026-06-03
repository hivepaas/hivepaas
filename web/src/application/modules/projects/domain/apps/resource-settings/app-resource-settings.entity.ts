export type AppResourceSettings = {
    reservations: ResourceReservations | null;
    limits: ResourceLimits | null;
    memory: ResourceMemory | null;
    ulimits: Ulimit[];
    capabilities: Capabilities | null;
    updateVer: number;
};

export type ResourceReservations = {
    cpus?: number;
    memory?: string;
    genericResources?: GenericResource[];
};

export type GenericResource = {
    kind: string;
    value: string;
};

export type ResourceLimits = {
    cpus?: number;
    memory?: string;
    pids?: number;
};

export type ResourceMemory = {
    swap?: string;
    swappiness?: number;
    shmSize?: string;
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
