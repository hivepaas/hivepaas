import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { SettingGithubApp } from "~/settings/domain";
import { GithubAppStatusBadge } from "~/settings/module-shared/components";

import { GithubAppEditCell, GithubAppMenuCell } from "./building-blocks";
import type { GithubAppTableScope } from "./github-app-table.types";

function createColumns(scope: GithubAppTableScope): ColumnDef<SettingGithubApp>[] {
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
                <GithubAppEditCell
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
            accessorKey: "organization",
            header: "Organization",
            enableSorting: true,
            cell: ({ row: { original } }) => original.organization || "-",
        },
        {
            accessorKey: "appId",
            header: "App ID",
            enableSorting: true,
            cell: ({ row: { original } }) => (original.appId ? original.appId : "-"),
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
                    <GithubAppStatusBadge status={original.status} />
                    {scope.type === "project" && original.inherited && (
                        <Badge className="bg-purple-500 text-white">Inherited</Badge>
                    )}
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
                <GithubAppMenuCell
                    scope={scope}
                    githubApp={original}
                />
            ),
            meta: {
                align: "right",
            },
        },
    ];
}

export const GithubAppTableDefs = Object.freeze({
    columns: createColumns,
});
