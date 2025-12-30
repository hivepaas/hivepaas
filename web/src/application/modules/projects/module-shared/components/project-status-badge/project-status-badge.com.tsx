import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { EProjectStatus } from "~/projects/module-shared/enums";

function View({ status }: Props) {
    const statusMap: Record<EProjectStatus, string> = {
        [EProjectStatus.Active]: "Active",
        [EProjectStatus.Locked]: "Locked",
        [EProjectStatus.Disabled]: "Disabled",
        [EProjectStatus.Deleting]: "Deleting",
    };

    const statusColorMap: Record<EProjectStatus, string> = {
        [EProjectStatus.Active]: "bg-green-500 text-white",
        [EProjectStatus.Locked]: "bg-orange-500 text-white",
        [EProjectStatus.Disabled]: "bg-red-500 text-white",
        [EProjectStatus.Deleting]: "bg-purple-500 text-white",
    };

    return (
        <Badge className={cn(statusColorMap[status] || "bg-primary text-primary-foreground", "h-6")}>
            {statusMap[status] || status}
        </Badge>
    );
}

interface Props {
    status: EProjectStatus;
}

export const ProjectStatusBadge = memo(View);
