import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";

import type { ProfileApiKey } from "@application/shared/entities/profile";

const columns: ColumnDef<ProfileApiKey>[] = [
    {
        accessorKey: "name",
        header: "Name",
    },
    {
        accessorKey: "key",
        header: "Key",
        cell: ({ row: { original } }) => {
            const { key } = original;
            // Mask the key, show first 8 chars and last 4 chars
            if (key.length > 12) {
                const masked = `${key.slice(0, 8)}...${key.slice(-4)}`;
                return <span className="font-mono">{masked}</span>;
            }
            return <span className="font-mono">{key}</span>;
        },
    },
    {
        header: "Access Actions",
        cell: ({ row: { original } }) => {
            const { accessAction } = original;
            const { read, write, delete: canDelete } = accessAction;
            const actions = [];
            if (read) actions.push("Read");
            if (write) actions.push("Write");
            if (canDelete) actions.push("Delete");
            return <span>{actions.length > 0 ? actions.join(", ") : "None"}</span>;
        },
    },
    {
        accessorKey: "expireAt",
        header: "Expires At",
        cell: ({ row: { original } }) => {
            const { expireAt } = original;
            try {
                return format(expireAt, "yyyy-MM-dd HH:mm:ss");
            } catch {
                return "-";
            }
        },
    },
    {
        accessorKey: "status",
        header: "Status",
        cell: ({ row: { original } }) => {
            const { status } = original;
            return (
                <span
                    className={`inline-flex items-center rounded-full px-2 py-1 text-xs font-medium ${
                        status === "active"
                            ? "bg-green-100 text-green-800"
                            : status === "expired"
                              ? "bg-red-100 text-red-800"
                              : "bg-gray-100 text-gray-800"
                    }`}
                >
                    {status}
                </span>
            );
        },
        meta: {
            align: "center",
            titleAlign: "center",
        },
    },
];

export const ProfileApiKeysTableDefs = Object.freeze({
    columns,
});
