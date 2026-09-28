import { memo } from "react";

import type { AppScheduledJob } from "~/projects/domain";
import { EAppScheduledJobType } from "~/projects/module-shared/enums";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

/** A job's runs; a data backup's snapshots too. */
function View({ projectId, env, appId, scheduledJob }: Props) {
    return (
        <div className="flex flex-col items-start gap-1">
            <AppLink.Basic
                className="text-sm font-medium text-link underline-offset-4"
                to={`${ROUTE.projects.single.apps.single.tasks.$route(projectId, env, appId)}?targetId=${scheduledJob.id}`}
            >
                View Runs
            </AppLink.Basic>
            {scheduledJob.jobType === EAppScheduledJobType.DataBackup && (
                <AppLink.Basic
                    className="text-sm font-medium text-link underline-offset-4"
                    to={`${ROUTE.projects.single.apps.single.configuration.backupSnapshots.$route(projectId, env, appId)}?tag=${encodeURIComponent(`hivepaas.job:${scheduledJob.id}`)}`}
                >
                    Snapshots
                </AppLink.Basic>
            )}
        </div>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    scheduledJob: AppScheduledJob;
}

export const ViewTasksCell = memo(View);
