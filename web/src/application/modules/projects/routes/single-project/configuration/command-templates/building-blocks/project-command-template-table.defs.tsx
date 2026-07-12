import { Badge } from "@components/ui/badge";
import { Checkbox } from "@components/ui/checkbox";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { ProjectCommandTemplate } from "~/projects/domain";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { ECommandTemplateKind } from "@application/shared/enums";

import { ProjectCommandTemplateEditCell } from "./project-command-template-edit-cell.com";
import { ProjectCommandTemplateMenuCell } from "./project-command-template-menu-cell.com";

function getKindClassName(kind: ProjectCommandTemplate["kind"]): string {
    switch (kind) {
        case ECommandTemplateKind.Database:
            return "bg-emerald-300 text-white";
        case ECommandTemplateKind.Backup:
            return "bg-blue-300 text-white";
        case ECommandTemplateKind.Deployment:
            return "bg-purple-300 text-white";
        case ECommandTemplateKind.Diagnostics:
            return "bg-amber-300 text-white";
        default:
            return "bg-slate-300 text-white";
    }
}

function createColumns(projectId: string): ColumnDef<ProjectCommandTemplate>[] {
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
                <ProjectCommandTemplateEditCell
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
            accessorKey: "kind",
            header: "Type",
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => (
                <div className="flex justify-center">
                    <Badge className={getKindClassName(original.kind)}>{original.kind}</Badge>
                </div>
            ),
        },
        {
            id: "script",
            header: "Script",
            enableSorting: false,
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => (
                <Checkbox
                    checked={original.script !== ""}
                    disabled
                    aria-label="Script mode"
                />
            ),
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
                <ProjectCommandTemplateMenuCell
                    projectId={projectId}
                    commandTemplate={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const ProjectCommandTemplateTableDefs = Object.freeze({
    columns: createColumns,
});
