import { memo } from "react";

import { Badge } from "@components/ui/badge";
import { cn } from "@lib/utils";

import { ESettingStatus } from "@application/shared/enums";

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
    const label =
        status === ESettingStatus.Active
            ? "Active"
            : status === ESettingStatus.Disabled
              ? "Disabled"
              : status === ESettingStatus.Expired
                ? "Expired"
                : status === ESettingStatus.Pending
                  ? "Pending"
                  : status === ESettingStatus.Missing
                    ? "Missing"
                    : formatStatusLabel(status);

    return (
        <Badge
            className={cn(
                status === ESettingStatus.Active && "bg-green-500 text-white",
                status === ESettingStatus.Disabled && "bg-red-600 text-white",
                status === ESettingStatus.Expired && "bg-red-500 text-white",
                status === ESettingStatus.Pending && "bg-amber-500 text-white",
                status === ESettingStatus.Missing && "bg-red-600 text-white",
            )}
        >
            {label}
        </Badge>
    );
}

interface Props {
    status: OpenApiConstant<ESettingStatus>;
}

export const GithubAppStatusBadge = memo(View);
