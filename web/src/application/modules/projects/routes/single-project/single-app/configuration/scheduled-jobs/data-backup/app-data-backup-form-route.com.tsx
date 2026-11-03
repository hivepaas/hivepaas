import { useState } from "react";

import { toast } from "sonner";
import { AppScheduledJobsCommands } from "~/projects/data";
import type { AppScheduledJob } from "~/projects/domain";
import {
    DataBackupForm,
    type DataBackupFormOutput,
    mapDataBackupFormToPayload,
} from "~/projects/module-shared/components";

import { AppLoader, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

export function AppDataBackupFormRoute({ mode, projectId, env, appId, scheduledJob }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();

    function navigateToList() {
        navigate.modules(ROUTE.projects.single.apps.single.configuration.scheduledJobs.$route(projectId, env, appId), {
            ignorePrevPath: true,
        });
    }

    const { mutate: createOne, isPending: isCreating } = AppScheduledJobsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Data backup created successfully");
            navigateToList();
        },
    });
    const { mutate: updateOne, isPending: isUpdating } = AppScheduledJobsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Data backup updated successfully");
            navigateToList();
        },
    });

    function onSubmit(values: DataBackupFormOutput) {
        if (!canWrite) {
            return;
        }

        const payload = mapDataBackupFormToPayload(values, appId);
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
            <RouteFormHeader title={mode === "create" ? "Create Data Backup" : "Edit Data Backup"} />

            {mode === "edit" && !scheduledJob ? (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            ) : (
                <DataBackupForm
                    projectId={projectId}
                    env={env}
                    appId={appId}
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
    /** The data backup being edited. */
    scheduledJob?: AppScheduledJob;
}
