import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { AppConfigFile } from "~/projects/domain";
import { ProjectSecretStatusBadge } from "~/projects/module-shared/components";

import { MenuCell } from "./building-blocks";

function createColumns(projectId: string, appId: string): ColumnDef<AppConfigFile>[] {
    return [
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
            header: "Mountpoint",
            cell: ({ row: { original } }) => {
                return original.swarmRef?.file?.name ?? "-";
            },
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
            header: "Actions",
            cell: ({ row: { original } }) => {
                if (original.inherited) {
                    return null;
                }

                return (
                    <MenuCell
                        projectId={projectId}
                        appId={appId}
                        configFile={original}
                    />
                );
            },
        },
    ];
}

export const AppConfigFilesTableDefs = Object.freeze({
    columns: createColumns,
});
