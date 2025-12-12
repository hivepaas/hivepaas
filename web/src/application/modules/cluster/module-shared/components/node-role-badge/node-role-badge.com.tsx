import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { ENodeRole } from "~/cluster/module-shared/enums";

function View({ role, isLeader }: Props) {
    const roleColorMap: Record<ENodeRole, string> = {
        [ENodeRole.Manager]: "bg-primary text-white",
        [ENodeRole.Worker]: "bg-blue-500 text-white",
    };

    const roleMap: Record<ENodeRole, string> = {
        [ENodeRole.Manager]: "Manager",
        [ENodeRole.Worker]: "Worker",
    };

    // If role is manager and isLeader is true, show "Leader" instead
    const displayText = role === ENodeRole.Manager && isLeader ? "Leader" : roleMap[role];
    const displayColor = role === ENodeRole.Manager && isLeader ? "bg-purple-500 text-white" : roleColorMap[role];

    return (
        <Badge className={cn(displayColor || "bg-primary text-primary-foreground", "h-6")}>{displayText || role}</Badge>
    );
}

interface Props {
    role: ENodeRole;
    isLeader: boolean;
}

export const NodeRoleBadge = memo(View);
