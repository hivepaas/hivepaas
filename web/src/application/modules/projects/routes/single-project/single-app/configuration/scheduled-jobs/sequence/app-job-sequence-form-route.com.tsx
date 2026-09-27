import { useMemo, useState } from "react";

import { toast } from "sonner";
import { APP_CONFIGURATION_QUERY_OPTIONS, AppScheduledJobsCommands, AppScheduledJobsQueries } from "~/projects/data";
import type { AppScheduledJob } from "~/projects/domain";
import {
    type JobSequenceCandidate,
    JobSequenceForm,
    type JobSequenceFormOutput,
    mapJobSequenceFormToPayload,
} from "~/projects/module-shared/components";
import { EAppScheduledJobType } from "~/projects/module-shared/enums";

import { AppLoader, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

const LIST_ALL_PAGE = { page: 1, size: 1000 };

/** An app's sequence runs the app's own jobs, never another sequence. */
function toSequenceCandidates(jobs: AppScheduledJob[]): JobSequenceCandidate[] {
    return jobs
        .filter(job => job.jobType !== EAppScheduledJobType.JobSequence)
        .map(job => ({ id: job.id, name: job.name, appName: "", disabled: job.status !== "active" }));
}

export function AppJobSequenceFormRoute({ mode, projectId, env, appId, scheduledJob }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();

    const { data: jobsData, isFetching: isLoadingCandidates } = AppScheduledJobsQueries.useFindManyPaginated(
        { projectID: projectId, env, appID: appId, pagination: LIST_ALL_PAGE },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );
    const candidates = useMemo(() => toSequenceCandidates(jobsData?.data ?? []), [jobsData]);

    function navigateToList() {
        navigate.modules(ROUTE.projects.single.apps.single.configuration.scheduledJobs.$route(projectId, env, appId), {
            ignorePrevPath: true,
        });
    }

    const { mutate: createOne, isPending: isCreating } = AppScheduledJobsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Job sequence created successfully");
            navigateToList();
        },
    });
    const { mutate: updateOne, isPending: isUpdating } = AppScheduledJobsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Job sequence updated successfully");
            navigateToList();
        },
    });

    function onSubmit(values: JobSequenceFormOutput) {
        if (!canWrite) {
            return;
        }

        const payload = mapJobSequenceFormToPayload(values);
        if (mode === "edit" && scheduledJob) {
            updateOne({
                projectID: projectId,
                env,
                appID: appId,
                scheduledJobID: scheduledJob.id,
                payload: { ...payload, updateVer: scheduledJob.updateVer },
            });
            return;
        }

        createOne({ projectID: projectId, env, appID: appId, payload });
    }

    function handleClose() {
        if (isCreating || isUpdating) {
            return;
        }

        if (canWrite && hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) {
            return;
        }

        navigateToList();
    }

    return (
        <div className="flex w-full flex-col">
            <RouteFormHeader title={mode === "create" ? "Create Job Sequence" : "Edit Job Sequence"} />

            {mode === "edit" && !scheduledJob ? (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            ) : (
                <JobSequenceForm
                    projectId={projectId}
                    env={env}
                    candidates={candidates}
                    isLoadingCandidates={isLoadingCandidates}
                    isPending={isCreating || isUpdating}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    initialValues={scheduledJob}
                    readOnly={!canWrite}
                    stickyActions
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

interface Props {
    mode: "create" | "edit";
    projectId: string;
    env: string;
    appId: string;
    /** The sequence being edited. */
    scheduledJob?: AppScheduledJob;
}
