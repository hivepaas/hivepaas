import { useState } from "react";

import { toast } from "sonner";
import type { AppScheduledJobs_Upsert_Payload } from "~/projects/api/services";
import { AppScheduledJobsCommands, AppScheduledJobsQueries } from "~/projects/data";
import { CreateOrEditAppScheduledJobForm } from "~/projects/dialogs/create-or-edit-app-scheduled-job/form";
import { APP_SCHEDULED_JOB_COMMAND_MODE } from "~/projects/dialogs/create-or-edit-app-scheduled-job/schemas";
import type { CreateOrEditAppScheduledJobFormOutput } from "~/projects/dialogs/create-or-edit-app-scheduled-job/schemas";
import { EAppScheduledJobScheduleMode, EAppScheduledJobType } from "~/projects/module-shared/enums";

import { AppLoader, RouteFormHeader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

type AppScheduledJobFormRouteMode = "create" | "edit";

function hasText(value: string): boolean {
    return value.trim().length > 0;
}

function mapFormValuesToPayload(
    values: CreateOrEditAppScheduledJobFormOutput,
    appId: string,
): AppScheduledJobs_Upsert_Payload {
    const scheduleInterval =
        values.scheduleMode === EAppScheduledJobScheduleMode.Interval && hasText(values.scheduleInterval)
            ? values.scheduleInterval
            : undefined;
    const scheduleCronExpr =
        values.scheduleMode === EAppScheduledJobScheduleMode.Cron && hasText(values.scheduleCronExpr)
            ? values.scheduleCronExpr
            : undefined;

    return {
        availableInProjects: false,
        default: false,
        name: values.name,
        jobType: EAppScheduledJobType.ContainerCommand,
        schedule: {
            ...(scheduleInterval ? { interval: scheduleInterval } : {}),
            ...(scheduleCronExpr ? { cronExpr: scheduleCronExpr } : {}),
            ...(values.scheduleFrom ? { initialTime: values.scheduleFrom } : {}),
            ...(values.scheduleTo ? { endTime: values.scheduleTo } : {}),
        },
        app: {
            id: appId,
        },
        priority: values.priority,
        maxRetry: values.maxRetry ?? 0,
        ...(hasText(values.retryDelay) ? { retryDelay: values.retryDelay } : {}),
        ...(hasText(values.retryDelayIncr) ? { retryDelayIncr: values.retryDelayIncr } : {}),
        retryBackoff: values.retryBackoff,
        ...(hasText(values.retryDelayMax) ? { retryDelayMax: values.retryDelayMax } : {}),
        ...(hasText(values.timeout) ? { timeout: values.timeout } : {}),
        controlDisabled: !values.controlEnabled,
        command: {
            command: values.commandMode === APP_SCHEDULED_JOB_COMMAND_MODE.Command ? values.command : "",
            script: values.commandMode === APP_SCHEDULED_JOB_COMMAND_MODE.Script ? values.script : "",
            consoleSize: values.consoleSize,
            tty: values.tty,
            ...(hasText(values.runInShell) ? { runInShell: values.runInShell } : {}),
            ...(hasText(values.workingDir) ? { workingDir: values.workingDir } : {}),
            ...(values.envVars.length > 0
                ? {
                      envVars: values.envVars.map(envVar => ({
                          ...envVar,
                      })),
                  }
                : {}),
            ...(values.argGroups.length > 0 ? { argGroups: values.argGroups } : {}),
        },
        notification: {
            successUseDefault: values.notification.successUseDefault,
            ...(!values.notification.successUseDefault && values.notification.success
                ? { success: values.notification.success }
                : {}),
            failureUseDefault: values.notification.failureUseDefault,
            ...(!values.notification.failureUseDefault && values.notification.failure
                ? { failure: values.notification.failure }
                : {}),
        },
    };
}

export function AppScheduledJobFormRoute({ mode, projectId, appId, scheduledJobId }: Props) {
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const { navigate } = useAppNavigate();
    const isEditMode = mode === "edit";

    function navigateToList() {
        navigate.modules(ROUTE.projects.single.apps.single.configuration.scheduledJobs.$route(projectId, appId), {
            ignorePrevPath: true,
        });
    }

    const { data: detailData, isLoading: isDetailLoading } = AppScheduledJobsQueries.useFindOneById(
        {
            projectID: projectId,
            appID: appId,
            scheduledJobID: scheduledJobId ?? "",
        },
        {
            enabled: isEditMode && Boolean(scheduledJobId),
        },
    );
    const scheduledJob = detailData?.data;

    const { mutate: createScheduledJob, isPending: isCreating } = AppScheduledJobsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("App scheduled job created successfully");
            navigateToList();
        },
    });

    const { mutate: updateScheduledJob, isPending: isUpdating } = AppScheduledJobsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("App scheduled job updated successfully");
            navigateToList();
        },
    });

    function onSubmit(values: CreateOrEditAppScheduledJobFormOutput) {
        if (!canWrite) {
            return;
        }

        const payload = mapFormValuesToPayload(values, appId);

        if (isEditMode) {
            if (!scheduledJob || !scheduledJobId) {
                return;
            }

            updateScheduledJob({
                projectID: projectId,
                appID: appId,
                scheduledJobID: scheduledJobId,
                payload: {
                    ...payload,
                    default: scheduledJob.default,
                    updateVer: scheduledJob.updateVer,
                },
            });
            return;
        }

        createScheduledJob({
            projectID: projectId,
            appID: appId,
            payload,
        });
    }

    function handleClose(): void {
        if (isCreating || isUpdating) {
            return;
        }

        if (canWrite && hasChanges) {
            const userConfirmed = window.confirm("Are you sure you want to close without saving changes?");
            if (!userConfirmed) {
                return;
            }
        }

        setHasChanges(false);
        navigateToList();
    }

    const shouldRenderForm = mode === "create" || Boolean(scheduledJob);

    return (
        <div className="flex w-full flex-col">
            <RouteFormHeader title={mode === "create" ? "Create Scheduled Job" : "Edit Scheduled Job"} />

            {isEditMode && isDetailLoading && (
                <div className="flex min-h-[220px] items-center justify-center">
                    <AppLoader />
                </div>
            )}

            {!(isEditMode && isDetailLoading) && shouldRenderForm && (
                <CreateOrEditAppScheduledJobForm
                    projectId={projectId}
                    isPending={isCreating || isUpdating}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    initialValues={isEditMode ? scheduledJob : undefined}
                    readOnly={!canWrite}
                    stickyActions
                    onClose={handleClose}
                />
            )}
        </div>
    );
}

interface Props {
    mode: AppScheduledJobFormRouteMode;
    projectId: string;
    appId: string;
    scheduledJobId?: string;
}
