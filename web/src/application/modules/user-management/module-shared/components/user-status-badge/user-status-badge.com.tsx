import { memo } from "react";

import { cn } from "@/lib/utils";
import { Badge, statusClassName } from "@components/ui/badge";

import { EUserStatus } from "@application/shared/enums";

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
    const statusMap: Partial<Record<EUserStatus, string>> = {
        [EUserStatus.Active]: "Active",
        [EUserStatus.Pending]: "Pending",
        [EUserStatus.Disabled]: "Disabled",
        [EUserStatus.Missing]: "Missing",
    };

    const statusColorMap: Partial<Record<EUserStatus, string>> = {
        [EUserStatus.Active]: statusClassName("success"),
        [EUserStatus.Pending]: statusClassName("attention"),
        [EUserStatus.Disabled]: statusClassName("failure"),
        [EUserStatus.Missing]: statusClassName("failure"),
    };

    return (
        <Badge className={cn(statusColorMap[status as EUserStatus] ?? "bg-primary text-primary-foreground")}>
            {statusMap[status as EUserStatus] ?? formatStatusLabel(status)}
        </Badge>
    );
}

interface Props {
    status: OpenApiConstant<EUserStatus>;
}

export const UserStatusBadge = memo(View);
