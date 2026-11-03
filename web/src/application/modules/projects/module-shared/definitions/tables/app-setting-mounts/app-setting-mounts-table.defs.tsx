import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { LockIcon } from "lucide-react";
import type { AppSettingMount } from "~/projects/domain";
import { ProjectSecretStatusBadge } from "~/projects/module-shared/components";

import { EditCell, MenuCell, StateCell } from "./building-blocks";

const SOURCE_TYPE_LABELS: Record<string, string> = {
    "secret": "Secret",
    "config-file": "Config file",
    "ssl-cert": "SSL certificate",
    "ssh-key": "SSH key",
    "basic-auth": "Basic auth",
};

/**
 * lockedIds are the rows a managed database's engine made and relies on: they
 * are shown, not changed.
 */
function createColumns(
    projectId: string,
    env: string,
    appId: string,
    lockedIds: string[] = [],
): ColumnDef<AppSettingMount>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) =>
                lockedIds.includes(original.id) ? null : (
                    <EditCell
                        projectId={projectId}
                        env={env}
                        appId={appId}
                        settingMount={original}
                    />
                ),
            meta: { align: "center", titleAlign: "center" },
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
            header: "Source",
            cell: ({ row: { original } }) => (
                <div className="flex flex-col">
                    <span className="break-all">{original.source.name || original.source.id}</span>
                    <span className="text-xs text-muted-foreground">
                        {SOURCE_TYPE_LABELS[original.source.type] ?? original.source.type}
                    </span>
                </div>
            ),
        },
        {
            header: "Files",
            cell: ({ row: { original } }) => (
                <div className="flex flex-col gap-1">
                    {original.files.map(file => (
                        <div
                            key={file.part}
                            className="flex items-center gap-1 text-xs"
                        >
                            {file.gated && <LockIcon className="size-3 text-amber-600" />}
                            <span className="text-muted-foreground">{file.part}</span>
                            <code className="break-all">{file.path}</code>
                        </div>
                    ))}
                </div>
            ),
        },
        {
            header: "Status",
            cell: ({ row: { original } }) => (
                <div className="flex items-center justify-center gap-2">
                    <ProjectSecretStatusBadge status={original.status} />
                    {original.inheritable && <Badge tone="purple">Inheritable</Badge>}
                </div>
            ),
            meta: { align: "center", titleAlign: "center" },
        },
        {
            header: "In the container",
            cell: ({ row: { original } }) => <StateCell settingMount={original} />,
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) =>
                lockedIds.includes(original.id) ? null : (
                    <MenuCell
                        projectId={projectId}
                        env={env}
                        appId={appId}
                        settingMount={original}
                    />
                ),
            meta: { align: "right" },
        },
    ];
}

export const AppSettingMountsTableDefs = Object.freeze({
    columns: createColumns,
    sourceTypeLabels: SOURCE_TYPE_LABELS,
});
