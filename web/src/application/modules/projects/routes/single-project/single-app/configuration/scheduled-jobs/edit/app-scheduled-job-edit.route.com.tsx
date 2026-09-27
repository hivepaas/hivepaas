import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { APP_CONFIGURATION_QUERY_OPTIONS, AppScheduledJobsQueries } from "~/projects/data";
import { EAppScheduledJobType } from "~/projects/module-shared/enums";

import { AppLoader } from "@application/shared/components";

import { AppDataBackupFormRoute } from "../data-backup";
import { AppScheduledJobFormRoute } from "../form-route";
import { AppJobSequenceFormRoute } from "../sequence";

/** Opens the form the job's type needs: a sequence and a data backup have their own. */
export function AppScheduledJobEditRoute() {
    const {
        id: projectId,
        env,
        appId,
        scheduledJobId,
    } = useParams<{
        id: string;
        env: string;
        appId: string;
        scheduledJobId: string;
    }>();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");
    invariant(scheduledJobId, "scheduledJobId must be defined");

    const { data } = AppScheduledJobsQueries.useFindOneById(
        { projectID: projectId, env, appID: appId, scheduledJobID: scheduledJobId },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );
    const job = data?.data;

    if (!job) {
        return (
            <div className="flex min-h-[220px] items-center justify-center">
                <AppLoader />
            </div>
        );
    }

    if (job.jobType === EAppScheduledJobType.JobSequence) {
        return (
            <AppJobSequenceFormRoute
                mode="edit"
                projectId={projectId}
                env={env}
                appId={appId}
                scheduledJob={job}
            />
        );
    }

    if (job.jobType === EAppScheduledJobType.DataBackup) {
        return (
            <AppDataBackupFormRoute
                mode="edit"
                projectId={projectId}
                env={env}
                appId={appId}
                scheduledJob={job}
            />
        );
    }

    return (
        <AppScheduledJobFormRoute
            mode="edit"
            projectId={projectId}
            env={env}
            appId={appId}
            scheduledJobId={scheduledJobId}
        />
    );
}
