import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { AppConfigFile } from "~/projects/domain";
import { InheritableCell, ProjectSecretStatusBadge } from "~/projects/module-shared/components";

import { EditCell, MenuCell } from "./building-blocks";

/**
 * lockedIds are the rows a managed database's engine made and relies on: they
 * are shown, not changed.
 */
function createColumns(
    projectId: string,
    env: string,
    appId: string,
    lockedIds: string[] = [],
): ColumnDef<AppConfigFile>[] {
    return [
        {
            id: "view",
            accessorKey: "inherited",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => {
                if (original.inherited || lockedIds.includes(original.id)) {
                    return null;
                }

                return (
                    <EditCell
                        projectId={projectId}
                        env={env}
                        appId={appId}
                        configFile={original}
                    />
                );
            },
            meta: {
                align: "center",
                titleAlign: "center",
            },
        },
        {
            accessorKey: "name",
            header: "Name",
            cell: ({ row: { original } }) => (
                <div className="flex flex-wrap items-center gap-1.5 break-all">
                    {original.name}
                    {lockedIds.includes(original.id) && (
                        <Badge
                            tone="violet"
                            title="The managed database's engine made it, and relies on it"
                        >
                            Managed
                        </Badge>
                    )}
                </div>
            ),
        },
        {
            header: "Status",
            cell: ({ row: { original } }) => {
                const { status } = original;
                return (
                    <div className="flex items-center justify-center gap-2">
                        <ProjectSecretStatusBadge status={status} />
                        {original.inherited && <Badge tone="purple">Inherited</Badge>}
                    </div>
                );
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
            id: "inheritable",
            header: "In Previews",
            enableSorting: false,
            size: 110,
            // An inherited row's flag is about the project's or env's apps, not this app's previews.
            cell: ({ row: { original } }) =>
                original.inherited ? null : (
                    <InheritableCell
                        inheritable={original.inheritable}
                        target="previews"
                    />
                ),
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
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => {
                if (original.inherited || lockedIds.includes(original.id)) {
                    return null;
                }

                return (
                    <MenuCell
                        projectId={projectId}
                        env={env}
                        appId={appId}
                        configFile={original}
                    />
                );
            },
            meta: {
                align: "right",
            },
        },
    ];
}

export const AppConfigFilesTableDefs = Object.freeze({
    columns: createColumns,
});
