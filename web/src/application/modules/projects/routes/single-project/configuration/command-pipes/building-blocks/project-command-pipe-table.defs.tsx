import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { ProjectCommandPipe } from "~/projects/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { ProjectCommandPipeEditCell } from "./project-command-pipe-edit-cell.com";
import { ProjectCommandPipeMenuCell } from "./project-command-pipe-menu-cell.com";

function createColumns(projectId: string): ColumnDef<ProjectCommandPipe>[] {
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
                <ProjectCommandPipeEditCell
                    projectId={projectId}
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
            accessorKey: "status",
            header: "Status",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => (
                <div className="flex items-center justify-center gap-2">
                    <SettingStatusBadge status={original.status} />
                    {original.inherited && <Badge className="bg-purple-500 text-white">Inherited</Badge>}
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
                <ProjectCommandPipeMenuCell
                    projectId={projectId}
                    commandPipe={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const ProjectCommandPipeTableDefs = Object.freeze({
    columns: createColumns,
});
