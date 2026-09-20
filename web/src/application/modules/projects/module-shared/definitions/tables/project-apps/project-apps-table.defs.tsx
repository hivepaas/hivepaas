import { cn } from "@/lib/utils";
import { Avatar } from "@components/ui/avatar";
import { Badge } from "@components/ui/badge";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import { ChevronRight } from "lucide-react";
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
            cell: ({ row }) => {
                const app = row.original;
                const isChild = row.depth > 0;
                const canExpand = row.getCanExpand();
                const isExpanded = row.getIsExpanded();
                const subAppCount = row.subRows.length;
                const indentPadding = isChild ? `${(row.depth - 1) * 1.5 + 2.75}rem` : undefined;
                const branchLeft = isChild ? `${(row.depth - 1) * 1.5 + 0.625}rem` : undefined;

                return (
                    <div
                        className={cn("flex items-center gap-2 min-w-0", isChild && "relative")}
                        style={{ paddingLeft: indentPadding }}
                    >
                        {isChild && (
                            <div
                                aria-hidden="true"
                                className="absolute top-0 bottom-1/2 w-[26px] border-l-2 border-b-2 border-border/80 rounded-bl-sm pointer-events-none"
                                style={{ left: branchLeft }}
                            />
                        )}
                        {canExpand ? (
                            <button
                                type="button"
                                onClick={e => {
                                    e.stopPropagation();
                                    row.toggleExpanded();
                                }}
                                className="flex size-5 shrink-0 items-center justify-center rounded-md p-0 text-muted-foreground hover:text-foreground hover:bg-muted/80 transition-colors"
                                aria-label={isExpanded ? "Collapse sub-apps" : "Expand sub-apps"}
                            >
                                <ChevronRight
                                    className={cn(
                                        "size-3.5 transition-transform duration-200",
                                        isExpanded && "rotate-90",
                                    )}
                                />
                            </button>
                        ) : (
                            !isChild && <span className="w-5 shrink-0" />
                        )}
                        <Avatar
                            name={app.name}
                            src={app.photo}
                            className={cn("rounded-lg shrink-0", isChild ? "size-6" : "size-7")}
                        />
                        <span
                            className={cn(
                                "font-medium truncate",
                                isChild ? "text-xs text-foreground/90" : "text-sm text-foreground",
                            )}
                        >
                            {app.name}
                        </span>
                        {isChild && (
                            <Badge
                                variant="secondary"
                                className="text-[10px] h-4 px-1.5 py-0 font-normal text-muted-foreground bg-muted/60 shrink-0"
                            >
                                child
                            </Badge>
                        )}
                        {canExpand && (
                            <Badge
                                variant="outline"
                                className="text-[10px] h-4 px-1.5 py-0 font-normal text-muted-foreground border-border/80 shrink-0"
                            >
                                {subAppCount} {subAppCount === 1 ? "sub-app" : "sub-apps"}
                            </Badge>
                        )}
                    </div>
                );
            },
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
