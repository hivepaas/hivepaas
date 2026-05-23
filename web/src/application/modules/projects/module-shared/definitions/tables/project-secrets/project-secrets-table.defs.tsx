import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { ProjectSecret } from "~/projects/domain";
import { ProjectSecretStatusBadge } from "~/projects/module-shared/components";

import { EditCell, MenuCell } from "./building-blocks";

function createColumns(projectId: string): ColumnDef<ProjectSecret>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <EditCell
                    projectId={projectId}
                    secret={original}
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
        },
        {
            header: "Status",
            cell: ({ row: { original } }) => {
                const { status } = original;
                return <ProjectSecretStatusBadge status={status} />;
            },
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            header: "Type",
            cell: ({ row: { original } }) => {
                return original.base64 ? "binary" : "text";
            },
        },
        {
            accessorKey: "updatedAt",
            header: "Last Updated",
            cell: ({ row: { original } }) => {
                const updatedAt = original.updatedAt ?? original.createdAt;
                return format(updatedAt, "yyyy-MM-dd HH:mm:ss");
            },
        },
        {
            header: "Actions",
            cell: ({ row: { original } }) => {
                return (
                    <MenuCell
                        projectId={projectId}
                        secret={original}
                    />
                );
            },
        },
    ];
}

export const ProjectSecretsTableDefs = Object.freeze({
    columns: createColumns,
});
