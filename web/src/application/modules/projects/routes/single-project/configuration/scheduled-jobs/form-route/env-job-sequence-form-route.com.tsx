import { useMemo, useState } from "react";

import { toast } from "sonner";
import { EnvScheduledJobsCommands } from "~/projects/data/commands";
import { EnvScheduledJobsQueries, ProjectAppsQueries } from "~/projects/data/queries";
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

/** An env's sequence runs the jobs of the env's apps, never another sequence. */
export function EnvJobSequenceFormRoute({ mode, projectId, env, scheduledJob }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();

    const { data: jobsData, isFetching: isLoadingCandidates } = EnvScheduledJobsQueries.useFindManyPaginated({
        projectID: projectId,
        env,
        pagination: LIST_ALL_PAGE,
    });
    const { data: appsData } = ProjectAppsQueries.useFindManyPaginated({
        projectID: projectId,
        env,
        pagination: LIST_ALL_PAGE,
    });
    const triggerApps = useMemo(
        () => (appsData?.data ?? []).map(app => ({ id: app.id, name: app.name })),
        [appsData?.data],
    );

    const candidates = useMemo<JobSequenceCandidate[]>(
        () =>
            (jobsData?.data ?? [])
                .filter(job => job.scope === "app" && job.jobType !== EAppScheduledJobType.JobSequence)
                .map(job => ({
                    id: job.id,
                    name: job.name,
                    appName: job.ownerApp?.name ?? "",
                    disabled: job.status !== "active",
                })),
        [jobsData],
    );

    function navigateToList() {
        navigate.modules(ROUTE.projects.single.providerConfiguration.scheduledJobs.$route(projectId), {
            ignorePrevPath: true,
        });
    }

    const { mutate: createOne, isPending: isCreating } = EnvScheduledJobsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Job sequence created successfully");
            navigateToList();
        },
    });
    const { mutate: updateOne, isPending: isUpdating } = EnvScheduledJobsCommands.useUpdateOne({
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
                scheduledJobID: scheduledJob.id,
                payload: { ...payload, updateVer: scheduledJob.updateVer },
            });
            return;
        }

        createOne({ projectID: projectId, env, payload });
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
            <RouteFormHeader
                title={mode === "create" ? `Create Job Sequence in ${env}` : `Edit Job Sequence in ${env}`}
            />

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
                    triggerApps={triggerApps}
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
    scheduledJob?: AppScheduledJob;
}
