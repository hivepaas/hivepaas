import { Badge, type BadgeTone } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import type { SettingBackupRepo } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import type { BackupRepoTableScope } from "./backup-repo-table.types";
import { BackupRepoEditCell, BackupRepoMenuCell } from "./building-blocks";

function getEngineTone(kind?: string): BadgeTone {
    if (kind?.toLowerCase() === "kopia") {
        return "sky";
    }

    return "neutral";
}

function createColumns(scope: BackupRepoTableScope): ColumnDef<SettingBackupRepo>[] {
    return [
        {
            id: "view",
            accessorKey: "inherited",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <BackupRepoEditCell
                    scope={scope}
                    id={original.id}
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
            id: "engine",
            accessorFn: row => row.kind ?? row.engine ?? "",
            header: "Engine",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => {
                const engine = original.kind ?? original.engine ?? "-";
                return (
                    <div className="flex justify-center">
                        <Badge tone={getEngineTone(original.kind ?? original.engine)}>{engine}</Badge>
                    </div>
                );
            },
        },
        {
            accessorKey: "description",
            header: "Description",
            cell: ({ row: { original } }) => <div className="break-all">{original.description ?? "-"}</div>,
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
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => (
                <BackupRepoMenuCell
                    scope={scope}
                    backupRepo={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const BackupRepoTableDefs = Object.freeze({
    columns: createColumns,
});
