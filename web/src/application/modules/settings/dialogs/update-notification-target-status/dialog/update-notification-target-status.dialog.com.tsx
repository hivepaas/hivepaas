import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectNotificationCommands } from "~/projects/data/commands";
import { ProjectNotificationQueries } from "~/projects/data/queries";
import { NotificationCommands } from "~/settings/data/commands";
import { NotificationQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { UpdateNotificationTargetStatusForm } from "../form";
import { useUpdateNotificationTargetStatusDialogState } from "../hooks";
import type { UpdateNotificationTargetStatusFormOutput } from "../schemas";

export function UpdateNotificationTargetStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateNotificationTargetStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: updateSettingStatus, isPending: isUpdatingSetting } = NotificationCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Notification target status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateProjectStatus, isPending: isUpdatingProject } = ProjectNotificationCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Project notification target status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    const detailId = state.mode === "open" ? state.id : "";
    const settingDetailQuery = NotificationQueries.useFindOneById(
        { id: detailId },
        { enabled: state.mode === "open" && state.scope.type === "settings" },
    );
    const projectDetailQuery = ProjectNotificationQueries.useFindOneById(
        {
            projectID: state.mode === "open" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        { enabled: state.mode === "open" && state.scope.type === "project" },
    );
    const detailQuery =
        state.mode === "open" && state.scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const notificationTarget = detailQuery.data?.data;

    function onSubmit(values: UpdateNotificationTargetStatusFormOutput) {
        if (state.mode !== "open" || !notificationTarget) return;

        const payload = {
            updateVer: notificationTarget.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            availableInProjects: state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectStatus({ projectID: state.scope.projectId, id: notificationTarget.id, payload });
            return;
        }

        updateSettingStatus({ id: notificationTarget.id, payload });
    }

    function handleClose() {
        if (isPending) return;
        if (
            !readOnlyInherited &&
            hasChanges &&
            !window.confirm("Are you sure you want to close without saving changes?")
        )
            return;
        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const resolvedDialogOptions = dialogOptions ?? {};
    const readOnlyInherited = resolvedDialogOptions.readOnlyInherited === true;
    const dialogTitle = readOnlyInherited
        ? `${resolvedDialogOptions.entityTitle ?? "Notification Target"} Status`
        : "Change status";
    const isPending = isUpdatingSetting || isUpdatingProject;
    const showAvailableInProjects = state.mode === "open" && state.scope.type === "settings";
    const initialValues = notificationTarget
        ? {
              status:
                  notificationTarget.status === ESettingStatus.Disabled
                      ? ESettingStatus.Disabled
                      : ESettingStatus.Active,
              expireAt: notificationTarget.expireAt ?? undefined,
              availableInProjects: notificationTarget.availableInProjects ?? false,
              default: notificationTarget.default ?? false,
          }
        : undefined;
    const isDetailLoading = state.mode === "open" && detailQuery.isFetching;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>{dialogTitle}</DialogTitle>
                </DialogHeader>
                {isDetailLoading && <AppLoader />}
                {state.mode === "open" && !isDetailLoading && initialValues && (
                    <UpdateNotificationTargetStatusForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        showAvailableInProjects={showAvailableInProjects}
                        readOnlyInherited={readOnlyInherited}
                        onClose={handleClose}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
