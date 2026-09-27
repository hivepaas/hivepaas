import { Button } from "@components/ui/button";
import type { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import { EyeIcon } from "lucide-react";
import type { EnvScheduledJob } from "~/projects/domain";
import { formatJobSchedule } from "~/projects/module-shared/components";
import { ScheduledJobNameCell } from "~/projects/module-shared/definitions/tables/app-scheduled-jobs";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { EnvScheduledJobMenuCell } from "./env-scheduled-job-menu-cell.com";

/** Where a job is edited: the env's own here, an app's on its app's page. */
function editRoute(projectId: string, env: string, job: EnvScheduledJob): string {
    if (job.scope === "app" && job.ownerApp) {
        return ROUTE.projects.single.apps.single.configuration.scheduledJobs.edit.$route(
            projectId,
            env,
            job.ownerApp.id,
            job.id,
        );
    }

    return ROUTE.projects.single.providerConfiguration.scheduledJobs.edit.$route(projectId, job.id);
}

/** Where a job's runs are listed. */
function runsRoute(projectId: string, env: string, job: EnvScheduledJob): string {
    if (job.scope === "app" && job.ownerApp) {
        return `${ROUTE.projects.single.apps.single.tasks.$route(projectId, env, job.ownerApp.id)}?targetId=${job.id}`;
    }

    return `${ROUTE.projects.single.operations.tasks.$route(projectId)}?targetId=${job.id}`;
}

function createColumns(projectId: string, env: string): ColumnDef<EnvScheduledJob>[] {
    return [
        {
            id: "view",
            header: "",
            enableSorting: false,
            enableHiding: false,
            minSize: 56,
            size: 56,
            cell: ({ row: { original } }) => (
                <AppLink.Basic to={editRoute(projectId, env, original)}>
                    <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-link hover:opacity-50"
                    >
                        <EyeIcon className="size-5" />
                        <span className="sr-only">Edit scheduled job</span>
                    </Button>
                </AppLink.Basic>
            ),
            meta: { align: "center", titleAlign: "center" },
        },
        {
            id: "runs",
            header: "",
            enableSorting: false,
            minSize: 100,
            size: 100,
            cell: ({ row: { original } }) => (
                <AppLink.Basic
                    className="text-sm font-medium text-link underline-offset-4"
                    to={runsRoute(projectId, env, original)}
                >
                    View Runs
                </AppLink.Basic>
            ),
        },
        {
            accessorKey: "name",
            header: "Name",
            cell: ({ row: { original } }) => <ScheduledJobNameCell job={original} />,
        },
        {
            id: "owner",
            header: "Owner",
            enableSorting: false,
            cell: ({ row: { original } }) =>
                original.scope === "project-env" ? "Env" : (original.ownerApp?.name ?? "-"),
        },
        {
            accessorKey: "schedule",
            header: "Schedule",
            enableSorting: false,
            cell: ({ row: { original } }) => formatJobSchedule(original.schedule),
        },
        {
            accessorKey: "nextRuns",
            header: "Next Run",
            enableSorting: false,
            cell: ({ row: { original } }) =>
                original.nextRuns[0] ? format(original.nextRuns[0], "yyyy-MM-dd HH:mm:ss") : "-",
        },
        {
            accessorKey: "status",
            header: "Status",
            cell: ({ row: { original } }) => <SettingStatusBadge status={original.status} />,
            meta: { align: "center", titleAlign: "center" },
        },
        {
            id: "actions",
            header: "",
            enableSorting: false,
            cell: ({ row: { original } }) => (
                <EnvScheduledJobMenuCell
                    projectId={projectId}
                    env={env}
                    job={original}
                />
            ),
            meta: { align: "right" },
        },
    ];
}

export const EnvScheduledJobsTableDefs = Object.freeze({
    columns: createColumns,
});
