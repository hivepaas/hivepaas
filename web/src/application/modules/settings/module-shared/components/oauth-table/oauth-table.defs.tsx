import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingOAuth } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { OAuthEditCell, OAuthMenuCell } from "./building-blocks";

function createColumns(): ColumnDef<SettingOAuth>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => <OAuthEditCell id={original.id} />,
            meta: { align: "center", titleAlign: "center" },
        },
        {
            accessorKey: "name",
            header: "Name",
            enableSorting: true,
        },
        {
            accessorKey: "kind",
            header: "Provider",
            cell: ({ row: { original } }) => {
                if (!original.kind) return "-";
                return <Badge className="bg-sky-500 text-white">{original.kind}</Badge>;
            },
        },
        {
            accessorKey: "organization",
            header: "Organization",
            enableSorting: true,
        },
        {
            accessorKey: "status",
            header: "Status",
            meta: { align: "center", titleAlign: "center" },
            cell: ({ row: { original } }) => (
                <div className="flex items-center justify-center gap-2">
                    <SettingStatusBadge status={original.status} />
                </div>
            ),
        },
        {
            accessorKey: "expireAt",
            header: "Expire At",
            cell: ({ row: { original } }) => {
                if (!original.expireAt) return "-";
                return format(original.expireAt, "yyyy-MM-dd HH:mm:ss");
            },
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => <OAuthMenuCell oauth={original} />,
            meta: { align: "right" },
        },
    ];
}

export const OAuthTableDefs = Object.freeze({
    columns: createColumns,
});
