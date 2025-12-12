import type { ENodeRole } from "~/cluster/module-shared/enums";

export interface NodeBase {
    id: string;
    name: string;
    hostname: string;
    addr: string;
    role: ENodeRole;
}
