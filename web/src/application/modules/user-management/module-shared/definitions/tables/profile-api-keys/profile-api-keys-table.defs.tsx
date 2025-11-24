import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";

import type { ProfileApiKey } from "@application/shared/entities/profile";
import { EProfileApiKeyAction, EProfileApiKeyStatus } from "@application/shared/enums";

const columns: ColumnDef<ProfileApiKey>[] = [
    {
        accessorKey: "name",
        header: "Name",
    },
    {
        header: "Key ID",
        cell: ({ row: { original } }) => {
            const { keyId } = original;
            return <span className="font-mono">{keyId}</span>;
        },
    },
    {
        header: "Access Actions",
        cell: ({ row: { original } }) => {
            const { accessAction } = original;
            if (!accessAction) return "-";
            return (
                <Badge
                    variant="default"
                    className={
                        accessAction === EProfileApiKeyAction.Read
                            ? "bg-blue-500"
                            : accessAction === EProfileApiKeyAction.Write
                              ? "bg-orange-500"
                              : "bg-red-500"
                    }
                >
                    {accessAction.charAt(0).toUpperCase() + accessAction.slice(1)}
                </Badge>
            );
        },
    },
    {
        header: "Status",
        cell: ({ row: { original } }) => {
            const { status } = original;
            return (
                <Badge
                    variant="default"
                    className={status === EProfileApiKeyStatus.Active ? "bg-green-500" : "bg-red-500"}
                >
                    {status.charAt(0).toUpperCase() + status.slice(1)}
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
];

export const ProfileApiKeysTableDefs = Object.freeze({
    columns,
});
