import type { ENodeStatus } from "~/cluster/module-shared/enums";

export interface NodeBase {
    id: string;
    name: string;
    hostname: string;
    addr: string;
    status: ENodeStatus;
}
