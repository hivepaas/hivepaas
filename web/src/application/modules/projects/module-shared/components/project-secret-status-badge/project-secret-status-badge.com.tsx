import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { EProjectSecretStatus } from "~/projects/module-shared/enums";

function View({ status }: Props) {
    const statusMap: Record<EProjectSecretStatus, string> = {
        [EProjectSecretStatus.Active]: "Active",
        [EProjectSecretStatus.Pending]: "Pending",
        [EProjectSecretStatus.Disabled]: "Disabled",
        [EProjectSecretStatus.Expired]: "Expired",
    };

    const statusColorMap: Record<EProjectSecretStatus, string> = {
        [EProjectSecretStatus.Active]: "bg-green-500 text-white",
        [EProjectSecretStatus.Pending]: "bg-yellow-500 text-white",
        [EProjectSecretStatus.Disabled]: "bg-red-500 text-white",
        [EProjectSecretStatus.Expired]: "bg-gray-500 text-white",
    };

    return (
        <Badge className={cn(statusColorMap[status] || "bg-primary text-primary-foreground", "h-6")}>
            {statusMap[status] || status}
        </Badge>
    );
}

interface Props {
    status: EProjectSecretStatus;
}

export const ProjectSecretStatusBadge = memo(View);
