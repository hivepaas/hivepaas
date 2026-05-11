import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { AppSecret } from "~/projects/domain";
import { ProjectSecretStatusBadge } from "~/projects/module-shared/components";

import { MenuCell } from "./building-blocks";

function createColumns(projectId: string, appId: string): ColumnDef<AppSecret>[] {
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
            meta: {
                align: "center",
                titleAlign: "center",
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
                // Inherited secrets are read-only
                if (original.inherited) {
                    return null;
                }

                return (
                    <MenuCell
                        projectId={projectId}
                        appId={appId}
                        secret={original}
                    />
                );
            },
        },
    ];
}

export const AppSecretsTableDefs = Object.freeze({
    columns: createColumns,
});
