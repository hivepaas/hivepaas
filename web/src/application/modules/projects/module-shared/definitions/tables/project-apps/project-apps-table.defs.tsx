import { cn } from "@/lib/utils";
import { Avatar } from "@components/ui/avatar";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { ProjectAppDetails, ProjectEnvEntity } from "~/projects/domain";
import { ProjectAppStatusBadge, ProjectEnvBadge } from "~/projects/module-shared/components";
import {
    APP_REPLICAS_STATUS_DOT_CLASS,
    getAppReplicasStatusLabel,
    resolveAppReplicasStatus,
} from "~/projects/module-shared/utils";

import { ActionsCell, MenuCell } from "./building-blocks";

const centerMeta = {
    align: "center",
    titleAlign: "center",
} as const;

function createColumns(projectId: string, projectEnvs: readonly ProjectEnvEntity[]): ColumnDef<ProjectAppDetails>[] {
    return [
        {
            id: "actions",
            header: "",
            minSize: 80,
            size: 80,
            meta: centerMeta,
            enableSorting: false,
            cell: ({ row: { original } }) => {
                return (
                    <ActionsCell
                        projectId={projectId}
                        env={original.env}
                        appId={original.id}
                    />
                );
            },
        },
        {
            accessorKey: "name",
            header: "Name",
            cell: ({ row: { original } }) => (
                <div className="flex items-center gap-2">
                    <Avatar
                        name={original.name}
                        src={original.photo}
                        className="rounded-lg"
                    />
                    <span>{original.name}</span>
                </div>
            ),
        },
        {
            accessorKey: "key",
            header: "Key",
        },
        {
            header: "Replicas",
            meta: centerMeta,
            cell: ({ row: { original } }) => {
                const { stats } = original;
                const runningTasks = stats?.runningTasks ?? 0;
                const desiredTasks = stats?.desiredTasks ?? 0;

                if (!stats || desiredTasks === 0) {
                    return <span className="text-muted-foreground">-</span>;
                }

                const replicasStatus = resolveAppReplicasStatus(runningTasks, desiredTasks);
                const replicasLabel = getAppReplicasStatusLabel(runningTasks, desiredTasks);

                return (
                    <div
                        className="flex items-center justify-center gap-2"
                        title={replicasLabel}
                    >
                        <span>
                            {runningTasks}/{desiredTasks}
                        </span>
                        <span
                            className={cn("size-2 rounded-full", APP_REPLICAS_STATUS_DOT_CLASS[replicasStatus])}
                            role="img"
                            aria-label={replicasLabel}
                        />
                    </div>
                );
            },
        },
        {
            header: "Env",
            meta: centerMeta,
            cell: ({ row: { original } }) => {
                if (!original.env) {
                    return <span className="text-muted-foreground">-</span>;
                }

                const projectEnv = projectEnvs.find(env => env.name === original.env);

                return (
                    <ProjectEnvBadge
                        name={original.env}
                        className="items-center"
                        color={projectEnv?.color}
                    />
                );
            },
        },
        {
            header: "Status",
            meta: centerMeta,
            cell: ({ row: { original } }) => {
                const { status } = original;
                return <ProjectAppStatusBadge status={status} />;
            },
        },
        {
            accessorKey: "updatedAt",
            header: "Last Updated",
            meta: centerMeta,
            cell: ({ row: { original } }) => {
                const updatedAt = original.updatedAt ?? original.createdAt;
                return format(updatedAt, "yyyy-MM-dd HH:mm:ss");
            },
        },
        {
            id: "menu",
            header: "",
            minSize: 56,
            size: 56,
            meta: centerMeta,
            enableSorting: false,
            cell: ({ row: { original } }) => {
                return (
                    <MenuCell
                        projectId={projectId}
                        appId={original.id}
                        appEnv={original.env}
                    />
                );
            },
        },
    ];
}

export const ProjectAppsTableDefs = Object.freeze({
    columns: createColumns,
});
