import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge, statusClassName } from "@components/ui/badge";
import { ENodeState } from "~/cluster/module-shared/enums";

function View({ state }: Props) {
    const stateMap: Record<ENodeState, string> = {
        [ENodeState.Unknown]: "Unknown",
        [ENodeState.Down]: "Down",
        [ENodeState.Ready]: "Ready",
        [ENodeState.Disconnected]: "Disconnected",
    };

    const stateColorMap: Record<ENodeState, string> = {
        [ENodeState.Unknown]: statusClassName("attention"),
        [ENodeState.Down]: statusClassName("failure"),
        [ENodeState.Ready]: statusClassName("success"),
        [ENodeState.Disconnected]: statusClassName("inactive"),
    };

    return (
        <Badge className={cn(stateColorMap[state] || "bg-primary text-primary-foreground", "h-6")}>
            {stateMap[state] || state}
        </Badge>
    );
}

interface Props {
    state: ENodeState;
}

export const NodeStateBadge = memo(View);
