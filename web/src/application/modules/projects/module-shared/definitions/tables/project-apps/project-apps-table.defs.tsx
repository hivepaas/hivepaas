import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { ProjectAppDetails } from "~/projects/domain";
import { ProjectAppStatusBadge } from "~/projects/module-shared/components";

import { ActionsCell, MenuCell } from "./building-blocks";

function createColumns(projectId: string): ColumnDef<ProjectAppDetails>[] {
    return [
        {
            id: "actions",
            header: "",
            minSize: 80,
            size: 80,
            meta: {
                align: "center",
                titleAlign: "center",
            },
            cell: ({ row: { original } }) => {
                return (
                    <ActionsCell
                        projectId={projectId}
                        appId={original.id}
                    />
                );
            },
        },
        {
            accessorKey: "name",
            header: "Name",
        },
        {
            accessorKey: "key",
            header: "Key",
        },
        {
            header: "Status",
            cell: ({ row: { original } }) => {
                const { status } = original;
                return <ProjectAppStatusBadge status={status} />;
            },
            meta: {
                align: "center",
                titleAlign: "center",
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
                        app={original}
                    />
                );
            },
        },
    ];
}

export const ProjectAppsTableDefs = Object.freeze({
    columns: createColumns,
});
