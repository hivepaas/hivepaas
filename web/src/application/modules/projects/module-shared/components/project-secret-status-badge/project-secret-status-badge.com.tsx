import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { EProjectSecretStatus } from "~/projects/module-shared/enums";

import type { OpenApiConstant } from "@infrastructure/api";

function formatStatusLabel(status: string) {
    if (!status.trim()) {
        return "-";
    }

    return status
        .replace(/[_-]+/g, " ")
        .replace(/\s+/g, " ")
        .trim()
        .replace(/\b\w/g, char => char.toUpperCase());
}

function View({ status }: Props) {
    const statusMap: Partial<Record<EProjectSecretStatus, string>> = {
        [EProjectSecretStatus.Active]: "Active",
        [EProjectSecretStatus.Pending]: "Pending",
        [EProjectSecretStatus.Disabled]: "Disabled",
        [EProjectSecretStatus.Expired]: "Expired",
        [EProjectSecretStatus.Missing]: "Missing",
    };

    const statusColorMap: Partial<Record<EProjectSecretStatus, string>> = {
        [EProjectSecretStatus.Active]: "bg-green-500 text-white",
        [EProjectSecretStatus.Pending]: "bg-yellow-500 text-white",
        [EProjectSecretStatus.Disabled]: "bg-red-500 text-white",
        [EProjectSecretStatus.Expired]: "bg-gray-500 text-white",
        [EProjectSecretStatus.Missing]: "bg-red-500 text-white",
    };

    return (
        <Badge
            className={cn(statusColorMap[status as EProjectSecretStatus] ?? "bg-primary text-primary-foreground", "h-6")}
        >
            {statusMap[status as EProjectSecretStatus] ?? formatStatusLabel(status)}
        </Badge>
    );
}

interface Props {
    status: OpenApiConstant<EProjectSecretStatus>;
}

export const ProjectSecretStatusBadge = memo(View);
