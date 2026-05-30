import React, { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import type { AppHealthChecks_Upsert_Payload } from "~/projects/api/services";
import { AppHealthChecksCommands } from "~/projects/data/commands";
import type { AppHealthCheckREST } from "~/projects/domain";
import {
    EAppHealthCheckRestMethod,
    EAppHealthCheckReturnBodyMode,
    EAppHealthCheckType,
} from "~/projects/module-shared/enums";

import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { CreateOrEditAppHealthCheckForm } from "../form";
import { useCreateOrEditAppHealthCheckDialogState } from "../hooks";
import type { CreateOrEditAppHealthCheckFormOutput } from "../schemas";

const BODY_METHODS = [EAppHealthCheckRestMethod.POST, EAppHealthCheckRestMethod.PUT] as const;

function getRestPayload(values: CreateOrEditAppHealthCheckFormOutput): AppHealthCheckREST | null {
    if (values.healthcheckType !== EAppHealthCheckType.REST) {
        return null;
    }

    const returnText =
        values.rest.returnBodyMode === EAppHealthCheckReturnBodyMode.Text
            ? {
                  exact: values.rest.textExact,
                  regex: values.rest.textRegex,
              }
            : null;
    const returnJSON =
        values.rest.returnBodyMode === EAppHealthCheckReturnBodyMode.JSON
            ? {
                  exact: values.rest.jsonExact,
                  contain: values.rest.jsonContain,
              }
            : null;

    return {
        url: values.rest.url,
        method: values.rest.method,
        contentType: values.rest.contentType,
        body: BODY_METHODS.includes(values.rest.method as (typeof BODY_METHODS)[number]) ? values.rest.body : "",
        returnCode: values.rest.returnCode,
        returnText,
        returnJSON,
    };
}

function mapFormValuesToPayload(values: CreateOrEditAppHealthCheckFormOutput): AppHealthChecks_Upsert_Payload {
    return {
        availableInProjects: false,
        default: false,
        name: values.name,
        healthcheckType: values.healthcheckType,
        interval: values.interval,
        maxRetry: values.maxRetry,
        retryDelay: values.retryDelay,
        timeout: values.timeout,
        saveResultTasks: true,
        rest: getRestPayload(values),
        grpc: values.healthcheckType === EAppHealthCheckType.GRPC ? values.grpc : null,
        notification: {
            successUseDefault: values.notification.successUseDefault,
            ...(!values.notification.successUseDefault && values.notification.success
                ? { success: values.notification.success }
                : {}),
            failureUseDefault: values.notification.failureUseDefault,
            ...(!values.notification.failureUseDefault && values.notification.failure
                ? { failure: values.notification.failure }
                : {}),
            minSendInterval: values.notification.minSendInterval,
        },
    };
}

export function CreateOrEditAppHealthCheckDialog() {
    const { state, props: dialogOptions, ...actions } = useCreateOrEditAppHealthCheckDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const { mode } = state;
    const open = mode !== "closed";
    const isEditMode = mode === "edit";
    const projectId = mode !== "closed" ? state.projectId : null;
    const appId = mode !== "closed" ? state.appId : null;

    const { mutate: createHealthCheck, isPending: isCreating } = AppHealthChecksCommands.useCreateOne({
        onSuccess: () => {
            toast.success("App health check created successfully");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    const { mutate: updateHealthCheck, isPending: isUpdating } = AppHealthChecksCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("App health check updated successfully");
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

    function onSubmit(values: CreateOrEditAppHealthCheckFormOutput) {
        if (!canWrite || !projectId || !appId) {
            return;
        }

        const payload = mapFormValuesToPayload(values);

        if (state.mode === "edit") {
            updateHealthCheck({
                projectID: projectId,
                appID: appId,
                healthCheckID: state.healthCheck.id,
                payload: {
                    ...payload,
                    default: state.healthCheck.default,
                    saveResultTasks: state.healthCheck.saveResultTasks,
                    updateVer: state.healthCheck.updateVer,
                },
            });
            return;
        }

        if (state.mode === "open") {
            createHealthCheck({
                projectID: projectId,
                appID: appId,
                payload,
            });
        }
    }

    function handleClose(): void {
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
            <DialogContent className="min-w-[390px] w-[800px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Create or update a health check</DialogTitle>
                </DialogHeader>
                <CreateOrEditAppHealthCheckForm
                    projectId={projectId}
                    isPending={isCreating || isUpdating}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    initialValues={isEditMode ? state.healthCheck : undefined}
                    readOnly={!canWrite}
                />
            </DialogContent>
        </Dialog>
    );
}
