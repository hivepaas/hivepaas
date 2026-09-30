import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge, statusClassName } from "@components/ui/badge";
import { EProjectStatus } from "~/projects/module-shared/enums";

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
    const statusMap: Partial<Record<EProjectStatus, string>> = {
        [EProjectStatus.Active]: "Active",
        [EProjectStatus.Locked]: "Locked",
        [EProjectStatus.Disabled]: "Disabled",
        [EProjectStatus.Deleting]: "Deleting",
        [EProjectStatus.Missing]: "Missing",
    };

    const statusColorMap: Partial<Record<EProjectStatus, string>> = {
        [EProjectStatus.Active]: statusClassName("success"),
        [EProjectStatus.Locked]: statusClassName("attention"),
        [EProjectStatus.Disabled]: statusClassName("failure"),
        [EProjectStatus.Deleting]: statusClassName("progress"),
        [EProjectStatus.Missing]: statusClassName("failure"),
    };

    return (
        <Badge className={cn(statusColorMap[status as EProjectStatus] ?? "bg-primary text-primary-foreground", "h-6")}>
            {statusMap[status as EProjectStatus] ?? formatStatusLabel(status)}
        </Badge>
    );
}

interface Props {
    status: OpenApiConstant<EProjectStatus>;
}

export const ProjectStatusBadge = memo(View);
