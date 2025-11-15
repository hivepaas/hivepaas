import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";

import { EUserStatus } from "@application/shared/enums";

function View({ status }: Props) {
    const statusMap: Record<EUserStatus, string> = {
        [EUserStatus.Active]: "Active",
        [EUserStatus.Pending]: "Pending",
        [EUserStatus.Disabled]: "Disabled",
    };

    const statusColorMap: Record<EUserStatus, string> = {
        [EUserStatus.Active]: "bg-green-500 text-white",
        [EUserStatus.Pending]: "bg-orange-400 text-white",
        [EUserStatus.Disabled]: "bg-red-500 text-white",
    };

    return (
        <Badge className={cn(statusColorMap[status] || "bg-primary text-primary-foreground")}>
            {statusMap[status] || status}
        </Badge>
    );
}

interface Props {
    status: EUserStatus;
}

export const UserStatusBadge = memo(View);
