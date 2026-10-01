import { Badge, statusClassName } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";

import { CAPABILITY_IDS } from "@application/shared/constants";
import type { ProfileApiKey } from "@application/shared/entities/profile";
import { EProfileApiKeyStatus } from "@application/shared/enums";

import { ActionsCell, KeyIdCell } from "./building-blocks";

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

const columns: ColumnDef<ProfileApiKey>[] = [
    {
        accessorKey: "name",
        header: "Name",
    },
    {
        header: "Key ID",
        cell: ({ row: { original } }) => {
            const { keyId } = original;
            return <KeyIdCell keyId={keyId} />;
        },
    },
    {
        header: "Access Actions",
        cell: ({ row: { original } }) => {
            const { accessAction } = original;
            if (!accessAction) return "-";
            return (
                <div className="flex items-center gap-2">
                    {accessAction.read && <Badge tone="blue">Read</Badge>}
                    {accessAction.execute && <Badge tone="violet">Execute</Badge>}
                    {accessAction.write && <Badge tone="orange">Write</Badge>}
                    {accessAction.delete && <Badge tone="red">Delete</Badge>}
                    {original.capabilities.includes(CAPABILITY_IDS.SecretReveal) && (
                        <Badge tone="red">Reveal secrets</Badge>
                    )}
                </div>
            );
        },
    },
    {
        header: "Status",
        cell: ({ row: { original } }) => {
            const { status } = original;
            const statusColorMap: Partial<Record<EProfileApiKeyStatus, string>> = {
                [EProfileApiKeyStatus.Active]: statusClassName("success"),
                [EProfileApiKeyStatus.Disabled]: statusClassName("failure"),
                [EProfileApiKeyStatus.Expired]: statusClassName("failure"),
                [EProfileApiKeyStatus.Missing]: statusClassName("failure"),
            };
            const statusMap: Partial<Record<EProfileApiKeyStatus, string>> = {
                [EProfileApiKeyStatus.Active]: "Active",
                [EProfileApiKeyStatus.Disabled]: "Disabled",
                [EProfileApiKeyStatus.Expired]: "Expired",
                [EProfileApiKeyStatus.Missing]: "Missing",
            };
            return (
                <Badge
                    variant="default"
                    className={statusColorMap[status as EProfileApiKeyStatus]}
                >
                    {statusMap[status as EProfileApiKeyStatus] ?? formatStatusLabel(status)}
                </Badge>
            );
        },
        meta: {
            align: "center",
            titleAlign: "center",
        },
    },
    {
        accessorKey: "expireAt",
        header: "Expires At",
        cell: ({ row: { original } }) => {
            const { expireAt } = original;
            try {
                if (!expireAt) return "-";
                return format(expireAt, "yyyy-MM-dd HH:mm:ss");
            } catch {
                return "-";
            }
        },
    },
    {
        header: "Actions",
        cell: ({ row: { original } }) => {
            return <ActionsCell apiKey={original} />;
        },
    },
];

export const ProfileApiKeysTableDefs = Object.freeze({
    columns,
});
