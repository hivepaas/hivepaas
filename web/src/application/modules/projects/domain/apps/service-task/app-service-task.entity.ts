export interface AppServiceTaskNode {
    id: string;
    name: string;
    hostname: string;
    addr: string;
    role: string;
    isLeader: boolean;
}

export interface AppServiceTaskStatus {
    timestamp: Date | null;
    state: string;
    message: string;
    err: string;
    containerStatus: {
        containerId: string;
    } | null;
}

export interface AppServiceTask {
    id: string;
    slot: number;
    desiredState: string;
    node: AppServiceTaskNode | null;
    status: AppServiceTaskStatus | null;
}
