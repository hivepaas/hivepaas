import { memo } from "react";

import { Badge, type BadgeTone } from "@components/ui/badge";

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

const STATUS_TONES: Partial<Record<string, BadgeTone>> = {
    [ESettingStatus.Active]: "green",
    [ESettingStatus.Disabled]: "red",
    [ESettingStatus.Expired]: "red",
    [ESettingStatus.Pending]: "amber",
    [ESettingStatus.Missing]: "red",
};

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
            // A status is the one solid badge of a row. One this does not know keeps the default look.
            tone={STATUS_TONES[status]}
            appearance="solid"
        >
            {label}
        </Badge>
    );
}

interface Props {
    status: OpenApiConstant<ESettingStatus>;
}

export const SettingStatusBadge = memo(View);
