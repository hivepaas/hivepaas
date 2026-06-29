import React, { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import type { AppScheduledJobs_Upsert_Payload } from "~/projects/api/services";
import { AppScheduledJobsCommands, AppScheduledJobsQueries } from "~/projects/data";
import { EAppScheduledJobScheduleMode, EAppScheduledJobType } from "~/projects/module-shared/enums";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { CreateOrEditAppScheduledJobForm } from "../form";
import { useCreateOrEditAppScheduledJobDialogState } from "../hooks";
import { APP_SCHEDULED_JOB_COMMAND_MODE, type CreateOrEditAppScheduledJobFormOutput } from "../schemas";

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

export function CreateOrEditAppScheduledJobDialog() {
    const { state, props: dialogOptions, ...actions } = useCreateOrEditAppScheduledJobDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const { mode } = state;
    const open = mode !== "closed";
    const isEditMode = mode === "edit";
    const projectId = mode !== "closed" ? state.projectId : null;
    const appId = mode !== "closed" ? state.appId : null;
    const scheduledJobId = mode === "edit" ? state.scheduledJobId : "";

    const { data: detailData, isLoading: isDetailLoading } = AppScheduledJobsQueries.useFindOneById(
        {
            projectID: projectId ?? "",
            appID: appId ?? "",
            scheduledJobID: scheduledJobId,
        },
        {
            enabled: isEditMode && Boolean(projectId && appId && scheduledJobId),
        },
    );

    const { mutate: createScheduledJob, isPending: isCreating } = AppScheduledJobsCommands.useCreateOne({
        onSuccess: () => {
            toast.success("App scheduled job created successfully");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    const { mutate: updateScheduledJob, isPending: isUpdating } = AppScheduledJobsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("App scheduled job updated successfully");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    useEffect(() => {
        if (mode === "closed") {
            setHasChanges(false);
        }
    }, [mode]);

    function onSubmit(values: CreateOrEditAppScheduledJobFormOutput) {
        if (!canWrite || !projectId || !appId) {
            return;
        }

        const payload = mapFormValuesToPayload(values, appId);

        if (state.mode === "edit") {
            if (!detailData?.data) {
                return;
            }

            updateScheduledJob({
                projectID: projectId,
                appID: appId,
                scheduledJobID: state.scheduledJobId,
                payload: {
                    ...payload,
                    default: detailData.data.default,
                    updateVer: detailData.data.updateVer,
                },
            });
            return;
        }

        if (state.mode === "open") {
            createScheduledJob({
                projectID: projectId,
                appID: appId,
                payload,
            });
        }
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
        actions.close();
        dialogOptions?.onClose?.();
    }

    if (!projectId || !appId) {
        return null;
    }

    return (
        <Dialog
            open={open}
            onOpenChange={isOpen => {
                if (!isOpen) {
                    handleClose();
                }
            }}
        >
            <DialogFixedContent className="min-w-[390px] w-[860px]">
                <DialogHeader>
                    <DialogTitle>Create or update a scheduled job</DialogTitle>
                </DialogHeader>
                {isEditMode && isDetailLoading ? (
                    <DialogBody>
                        <AppLoader />
                    </DialogBody>
                ) : (
                    <CreateOrEditAppScheduledJobForm
                        projectId={projectId}
                        isPending={isCreating || isUpdating}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={isEditMode ? detailData?.data : undefined}
                        readOnly={!canWrite}
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
