import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { ENodeStatus } from "~/cluster/module-shared/enums";

function View({ status }: Props) {
    const statusMap: Record<ENodeStatus, string> = {
        [ENodeStatus.Unknown]: "Unknown",
        [ENodeStatus.Down]: "Down",
        [ENodeStatus.Ready]: "Ready",
        [ENodeStatus.Disconnected]: "Disconnected",
    };

    const statusColorMap: Record<ENodeStatus, string> = {
        [ENodeStatus.Unknown]: "bg-orange-500 text-white",
        [ENodeStatus.Down]: "bg-red-500 text-white",
        [ENodeStatus.Ready]: "bg-green-500 text-white",
        [ENodeStatus.Disconnected]: "bg-purple-500 text-white",
    };

    return (
        <Badge className={cn(statusColorMap[status] || "bg-primary text-primary-foreground", "h-6")}>
            {statusMap[status] || status}
        </Badge>
    );
}

interface Props {
    status: ENodeStatus;
}

export const NodeStatusBadge = memo(View);
