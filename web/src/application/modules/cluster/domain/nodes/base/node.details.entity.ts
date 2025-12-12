import type { ENodeAvailability, ENodeStatus } from "~/cluster/module-shared/enums";

import type { NodeBase } from "./node.base.entity";

export interface NodeDetails extends NodeBase {
    isLeader: boolean;
    labels: Record<string, string> | null;
    availability: ENodeAvailability;
    status: ENodeStatus;
    platform: {
        architecture: string;
        os: string;
    } | null;
    resources: {
        cpus: number;
        memoryMB: number;
    } | null;

    engineDesc: {
        engineVersion: string;
        labels: Record<string, string> | null;
        plugins?: { name: string; type: string }[];
    } | null;
    updateVer: number;
    createdAt: Date;
    updatedAt: Date | null;
}
