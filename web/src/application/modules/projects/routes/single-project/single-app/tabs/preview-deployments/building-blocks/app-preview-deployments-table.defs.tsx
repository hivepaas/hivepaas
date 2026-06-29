import { cn } from "@/lib/utils";
import type { ColumnDef } from "@tanstack/react-table";
import { Eye } from "lucide-react";
import type { ProjectAppDetails, ProjectEnvEntity } from "~/projects/domain";
import { ProjectAppStatusBadge, ProjectEnvBadge } from "~/projects/module-shared/components";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Button } from "@/components/ui";

import { getExternalPreviewUrl, getPreviewSubdomain } from "../utils";

const centerMeta = {
    align: "center",
    titleAlign: "center",
} as const;

function ReplicasCell({ stats }: Pick<ProjectAppDetails, "stats">) {
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
}

function SubdomainCell({ accessLinks }: Pick<ProjectAppDetails, "accessLinks">) {
    const accessLink = accessLinks[0];
    const subdomain = getPreviewSubdomain(accessLinks);

    if (!accessLink || !subdomain) {
        return <span className="text-muted-foreground">-</span>;
    }

    return (
        <a
            href={getExternalPreviewUrl(accessLink)}
            target="_blank"
            rel="noopener noreferrer"
            className="font-medium text-primary underline-offset-4 hover:underline"
        >
            {subdomain}
        </a>
    );
}

function createColumns(projectId: string, projectEnvs: readonly ProjectEnvEntity[]): ColumnDef<ProjectAppDetails>[] {
    return [
        {
            id: "view",
            header: "",
            minSize: 56,
            size: 56,
            enableSorting: false,
            meta: centerMeta,
            cell: ({ row: { original } }) => (
                <Button
                    variant="ghost"
                    size="icon-sm"
                    asChild
                >
                    <AppLink.Modules
                        to={ROUTE.projects.single.apps.single.configuration.general.$route(projectId, original.id)}
                        ignorePrevPath
                    >
                        <Eye className="size-4" />
                        <span className="sr-only">Open preview app</span>
                    </AppLink.Modules>
                </Button>
            ),
        },
        {
            accessorKey: "name",
            header: "Name",
        },
        {
            id: "subdomain",
            header: "Subdomain",
            enableSorting: false,
            cell: ({ row: { original } }) => <SubdomainCell accessLinks={original.accessLinks} />,
        },
        {
            accessorKey: "key",
            header: "Key",
        },
        {
            id: "replicas",
            header: "Replicas",
            enableSorting: false,
            meta: centerMeta,
            cell: ({ row: { original } }) => <ReplicasCell stats={original.stats} />,
        },
        {
            id: "env",
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
            accessorKey: "status",
            header: "Status",
            meta: centerMeta,
            cell: ({ row: { original } }) => <ProjectAppStatusBadge status={original.status} />,
        },
    ];
}

export const AppPreviewDeploymentsTableDefs = Object.freeze({
    columns: createColumns,
});
