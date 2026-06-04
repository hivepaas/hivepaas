import { cn } from "@/lib/utils";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import type { ProjectAppDetails, ProjectEnvEntity } from "~/projects/domain";
import { ProjectAppStatusBadge, ProjectEnvBadge } from "~/projects/module-shared/components";

import { ActionsCell } from "./building-blocks";

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
            header: "Replicas",
            meta: centerMeta,
            cell: ({ row: { original } }) => {
                const { stats } = original;
                const runningTasks = stats?.runningTasks ?? 0;
                const desiredTasks = stats?.desiredTasks ?? 0;

                if (!stats || desiredTasks === 0) {
                    return <span className="text-muted-foreground">-</span>;
                }

                const replicaDotClass = runningTasks < desiredTasks ? "bg-orange-500" : "bg-green-500";

                return (
                    <div className="flex items-center justify-center gap-2">
                        <span>
                            {runningTasks}/{desiredTasks}
                        </span>
                        <span className={cn("size-2 rounded-full", replicaDotClass)} />
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
    ];
}

export const ProjectAppsTableDefs = Object.freeze({
    columns: createColumns,
});
