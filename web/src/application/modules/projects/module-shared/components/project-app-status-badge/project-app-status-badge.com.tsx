import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

function View({ status }: Props) {
    const statusMap: Record<EProjectAppStatus, string> = {
        [EProjectAppStatus.Active]: "Active",
        [EProjectAppStatus.Locked]: "Locked",
        [EProjectAppStatus.Disabled]: "Disabled",
        [EProjectAppStatus.Deleting]: "Deleting",
    };

    const statusColorMap: Record<EProjectAppStatus, string> = {
        [EProjectAppStatus.Active]: "bg-green-500 text-white",
        [EProjectAppStatus.Locked]: "bg-orange-500 text-white",
        [EProjectAppStatus.Disabled]: "bg-red-500 text-white",
        [EProjectAppStatus.Deleting]: "bg-purple-500 text-white",
    };

    return (
        <Badge className={cn(statusColorMap[status] || "bg-primary text-primary-foreground", "h-6")}>
            {statusMap[status] || status}
        </Badge>
    );
}

interface Props {
    status: EProjectAppStatus;
}

export const ProjectAppStatusBadge = memo(View);
