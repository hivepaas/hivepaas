import React, { useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { AppScheduledJobsCommands, AppScheduledJobsQueries } from "~/projects/data";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { ESettingStatus } from "@application/shared/enums";
import { useConditionalModule } from "@application/shared/permissions";

import { UpdateAppScheduledJobStatusForm } from "../form";
import { useUpdateAppScheduledJobStatusDialogState } from "../hooks";
import type { UpdateAppScheduledJobStatusFormOutput } from "../schemas";

export function UpdateAppScheduledJobStatusDialog() {
    const { state, props: dialogOptions, ...actions } = useUpdateAppScheduledJobStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const open = state.mode !== "closed";
    const projectId = state.mode === "open" ? state.projectId : "";
    const appId = state.mode === "open" ? state.appId : "";
    const scheduledJobId = state.mode === "open" ? state.scheduledJobId : "";

    const { data, isLoading } = AppScheduledJobsQueries.useFindOneById(
        {
            projectID: projectId,
            appID: appId,
            scheduledJobID: scheduledJobId,
        },
        {
            enabled: state.mode === "open",
        },
    );

    const { mutate: updateStatus, isPending } = AppScheduledJobsCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("App scheduled job status updated successfully");
            actions.close();
            dialogOptions?.onSuccess?.();
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    function onSubmit(values: UpdateAppScheduledJobStatusFormOutput) {
        if (!canWrite || state.mode !== "open" || !data?.data) {
            return;
        }

        updateStatus({
            projectID: state.projectId,
            appID: state.appId,
            scheduledJobID: state.scheduledJobId,
            payload: {
                updateVer: data.data.updateVer,
                status: values.status,
                expireAt: values.expireAt ?? null,
                availableInProjects: false,
                default: values.default,
            },
        });
    }

    function handleClose(): void {
        if (isPending) {
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

    const initialValues = data?.data
        ? {
              status: data.data.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: data.data.expireAt,
              default: data.data.default,
          }
        : undefined;

    return (
        <Dialog
            open={open}
            onOpenChange={isOpen => {
                if (!isOpen) {
                    handleClose();
                }
            }}
        >
            <DialogContent className="min-w-[390px] w-[720px]">
                <DialogHeader>
                    <DialogTitle>Change status</DialogTitle>
                </DialogHeader>
                {isLoading ? (
                    <AppLoader />
                ) : (
                    <UpdateAppScheduledJobStatusForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        readOnly={!canWrite}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
