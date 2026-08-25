import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";

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
                    {accessAction.read && (
                        <Badge
                            variant="default"
                            className="bg-blue-500"
                        >
                            Read
                        </Badge>
                    )}
                    {accessAction.execute && (
                        <Badge
                            variant="default"
                            className="bg-violet-500"
                        >
                            Execute
                        </Badge>
                    )}
                    {accessAction.write && (
                        <Badge
                            variant="default"
                            className="bg-orange-500"
                        >
                            Write
                        </Badge>
                    )}
                    {accessAction.delete && (
                        <Badge
                            variant="default"
                            className="bg-red-500"
                        >
                            Delete
                        </Badge>
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
                [EProfileApiKeyStatus.Active]: "bg-green-500",
                [EProfileApiKeyStatus.Disabled]: "bg-red-500",
                [EProfileApiKeyStatus.Expired]: "bg-yellow-500",
                [EProfileApiKeyStatus.Missing]: "bg-red-500",
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
