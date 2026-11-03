import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge, statusClassName } from "@components/ui/badge";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

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
    const statusMap: Partial<Record<EProjectAppStatus, string>> = {
        [EProjectAppStatus.Active]: "Active",
        [EProjectAppStatus.Locked]: "Locked",
        [EProjectAppStatus.Disabled]: "Disabled",
        [EProjectAppStatus.Deleting]: "Deleting",
        [EProjectAppStatus.Missing]: "Missing",
    };

    const statusColorMap: Partial<Record<EProjectAppStatus, string>> = {
        [EProjectAppStatus.Active]: statusClassName("success"),
        [EProjectAppStatus.Locked]: statusClassName("attention"),
        [EProjectAppStatus.Disabled]: statusClassName("failure"),
        [EProjectAppStatus.Deleting]: statusClassName("progress"),
        [EProjectAppStatus.Missing]: statusClassName("failure"),
    };

    return (
        <Badge
            className={cn(statusColorMap[status as EProjectAppStatus] ?? "bg-primary text-primary-foreground", "h-6")}
        >
            {statusMap[status as EProjectAppStatus] ?? formatStatusLabel(status)}
        </Badge>
    );
}

interface Props {
    status: OpenApiConstant<EProjectAppStatus>;
}

export const ProjectAppStatusBadge = memo(View);
