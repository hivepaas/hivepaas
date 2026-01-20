import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { ProjectAppDetails } from "~/projects/domain";
import { ProjectAppStatusBadge } from "~/projects/module-shared/components";

import { MenuCell } from "./building-blocks";

function createColumns(projectId: string): ColumnDef<ProjectAppDetails>[] {
    return [
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
