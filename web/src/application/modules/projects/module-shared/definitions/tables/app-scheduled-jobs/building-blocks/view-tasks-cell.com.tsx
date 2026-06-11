import { memo } from "react";

import type { AppScheduledJob } from "~/projects/domain";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

function View({ projectId, appId, scheduledJob }: Props) {
    return (
        <AppLink.Basic
            className="text-sm font-medium text-primary underline-offset-4"
            to={ROUTE.projects.single.apps.single.scheduledJobTasks.$route(projectId, appId, scheduledJob.id)}
        >
            View Tasks
        </AppLink.Basic>
    );
}

interface Props {
    projectId: string;
    appId: string;
    scheduledJob: AppScheduledJob;
}

export const ViewTasksCell = memo(View);
