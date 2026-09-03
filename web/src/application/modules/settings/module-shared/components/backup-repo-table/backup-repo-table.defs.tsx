import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import type { SettingBackupRepo } from "~/settings/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import type { BackupRepoTableScope } from "./backup-repo-table.types";
import { BackupRepoEditCell, BackupRepoMenuCell } from "./building-blocks";

function getEngineBadgeClassName(kind?: string): string {
    if (kind?.toLowerCase() === "kopia") {
        return "bg-sky-500 text-white border-transparent hover:bg-sky-500/90";
    }

    return "bg-slate-500 text-white border-transparent hover:bg-slate-500/90";
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
                        <Badge className={getEngineBadgeClassName(original.kind ?? original.engine)}>{engine}</Badge>
                    </div>
                );
            },
        },
        {
            accessorKey: "description",
            header: "Description",
            cell: ({ row: { original } }) => original.description ?? "-",
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
                    {scope.type === "project" && original.inherited && (
                        <Badge className="bg-purple-500 text-white">Inherited</Badge>
                    )}
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
