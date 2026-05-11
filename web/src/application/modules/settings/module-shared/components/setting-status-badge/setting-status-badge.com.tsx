import { memo } from "react";

import { cn } from "@lib/utils";

import { ESettingStatus } from "@application/shared/enums";

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
                  : "-";

    return (
        <span
            className={cn(
                "inline-flex items-center rounded-md px-2 py-1 text-xs font-medium",
                status === ESettingStatus.Active && "bg-green-500 text-white",
                status === ESettingStatus.Disabled && "bg-muted text-muted-foreground",
                status === ESettingStatus.Expired && "bg-destructive text-destructive-foreground",
                status === ESettingStatus.Pending && "bg-amber-500 text-white",
            )}
        >
            {label}
        </span>
    );
}

interface Props {
    status: ESettingStatus;
}

export const SettingStatusBadge = memo(View);
