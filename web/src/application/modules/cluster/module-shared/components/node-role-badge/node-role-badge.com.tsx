import { memo } from "react";

import { Badge, type BadgeTone } from "@components/ui/badge";
import { ENodeRole } from "~/cluster/module-shared/enums";

function View({ role, isLeader }: Props) {
    const roleToneMap: Record<ENodeRole, BadgeTone> = {
        [ENodeRole.Manager]: "amber",
        [ENodeRole.Worker]: "blue",
    };

    const roleMap: Record<ENodeRole, string> = {
        [ENodeRole.Manager]: "Manager",
        [ENodeRole.Worker]: "Worker",
    };

    // If role is manager and isLeader is true, show "Leader" instead
    const displayText = role === ENodeRole.Manager && isLeader ? "Leader" : roleMap[role];
    const displayTone: BadgeTone = role === ENodeRole.Manager && isLeader ? "purple" : roleToneMap[role];

    return (
        <Badge
            tone={displayTone}
            className="h-6"
        >
            {displayText || role}
        </Badge>
    );
}

interface Props {
    role: ENodeRole;
    isLeader: boolean;
}

export const NodeRoleBadge = memo(View);
