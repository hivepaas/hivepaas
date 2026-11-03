import { Badge, type BadgeTone } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { ClusterVolume } from "~/cluster/domain";
import type { VolumeManagementScope } from "~/cluster/module-shared/types";
import { getClusterVolumeTypeLabel } from "~/cluster/module-shared/utils";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { ActionsCell, MenuCell } from "./building-blocks";

/** The color of a local volume's type; one not known is neutral. */
function volumeTypeTone(type: string): BadgeTone {
    switch (type.toLowerCase()) {
        case "bind":
            return "amber";
        case "nfs":
            return "blue";
        case "tmpfs":
            return "purple";
        default:
            return "neutral";
    }
}

function createColumns(scope: VolumeManagementScope): ColumnDef<ClusterVolume>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <ActionsCell
                    volume={original}
                    scope={scope}
                />
            ),
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            accessorKey: "name",
            header: "Name",
            enableSorting: true,
        },
        {
            accessorKey: "driver",
            header: "Driver",
            enableSorting: true,
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => (original.driver ? <Badge tone="sky">{original.driver}</Badge> : "-"),
        },
        {
            id: "volumeType",
            header: "Type",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => {
                const type = getClusterVolumeTypeLabel(original);

                return type === "-" ? type : <Badge tone={volumeTypeTone(type)}>{type}</Badge>;
            },
        },
        {
            accessorKey: "status",
            header: "Status",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => (
                <div className="flex items-center justify-center gap-2">
                    <SettingStatusBadge status={original.status} />
                    {scope.type === "project" && original.inherited && <Badge tone="purple">Inherited</Badge>}
                </div>
            ),
        },
        {
            accessorKey: "expireAt",
            header: "Expire At",
            cell: ({ row: { original } }) => {
                if (!original.expireAt) {
                    return "-";
                }

                return format(original.expireAt, "yyyy-MM-dd HH:mm:ss");
            },
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => (
                <MenuCell
                    volume={original}
                    scope={scope}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const VolumesTableDefs = Object.freeze({
    columns: createColumns,
});
