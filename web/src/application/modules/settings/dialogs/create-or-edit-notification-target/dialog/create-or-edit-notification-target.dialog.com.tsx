import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectNotificationCommands } from "~/projects/data/commands";
import { ProjectNotificationQueries } from "~/projects/data/queries";
import { NotificationCommands } from "~/settings/data/commands";
import { NotificationQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";

import { CreateOrEditNotificationTargetForm } from "../form";
import { useCreateOrEditNotificationTargetDialogState } from "../hooks";
import type { CreateOrEditNotificationTargetFormOutput } from "../schemas";

function splitCommaSeparated(value: string): string[] {
    return value
        .split(",")
        .map(item => item.trim())
        .filter(Boolean);
}

function joinAddresses(value?: string[]): string {
    return value?.join(", ") ?? "";
}

export function CreateOrEditNotificationTargetDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useCreateOrEditNotificationTargetDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: createSettingNotification, isPending: isCreatingSetting } = NotificationCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Notification target created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateSettingNotification, isPending: isUpdatingSetting } = NotificationCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Notification target updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: createProjectNotification, isPending: isCreatingProject } =
        ProjectNotificationCommands.useCreateOne({
            onSuccess: () => {
                toast.success("Project notification target created successfully");
                closeDialog();
                dialogOptions?.onSuccess?.();
            },
        });
    const { mutate: updateProjectNotification, isPending: isUpdatingProject } =
        ProjectNotificationCommands.useUpdateOne({
            onSuccess: () => {
                toast.success("Project notification target updated successfully");
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

    const detailId = state.mode === "edit" ? state.id : "";
    const settingDetailQuery = NotificationQueries.useFindOneById(
        { id: detailId },
        { enabled: state.mode === "edit" && state.scope.type === "settings" },
    );
    const projectDetailQuery = ProjectNotificationQueries.useFindOneById(
        {
            projectID: state.mode === "edit" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        { enabled: state.mode === "edit" && state.scope.type === "project" },
    );
    const detailQuery =
        state.mode === "edit" && state.scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const notificationTarget = detailQuery.data?.data;

    function createPayload(values: CreateOrEditNotificationTargetFormOutput) {
        const viaEmail =
            !values.emailEnabled && !values.senderEmailAccountId
                ? null
                : {
                      enabled: values.emailEnabled,
                      sender: { id: values.senderEmailAccountId },
                      toProjectMembers: values.notifyProjectMembers,
                      toProjectOwners: values.notifyProjectOwners,
                      toAllAdmins: values.notifyAdmins,
                      toAddresses: splitCommaSeparated(values.customAddresses),
                  };
        const viaSlack =
            !values.slackEnabled && !values.slackWebhookId
                ? null
                : {
                      enabled: values.slackEnabled,
                      webhook: { id: values.slackWebhookId },
                  };
        const viaDiscord =
            !values.discordEnabled && !values.discordWebhookId
                ? null
                : {
                      enabled: values.discordEnabled,
                      webhook: { id: values.discordWebhookId },
                  };

        return {
            availableInProjects:
                state.mode !== "closed" && state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            name: values.name,
            viaEmail,
            viaSlack,
            viaDiscord,
            minSendInterval: values.minSendInterval,
        };
    }

    function onSubmit(values: CreateOrEditNotificationTargetFormOutput) {
        if (state.mode === "closed") return;
        const payload = createPayload(values);

        if (state.mode === "edit" && notificationTarget) {
            const updatePayload = { ...payload, updateVer: notificationTarget.updateVer };
            if (state.scope.type === "project") {
                updateProjectNotification({
                    projectID: state.scope.projectId,
                    id: notificationTarget.id,
                    payload: updatePayload,
                });
                return;
            }
            updateSettingNotification({ id: notificationTarget.id, payload: updatePayload });
            return;
        }

        if (state.scope.type === "project") {
            createProjectNotification({ projectID: state.scope.projectId, payload });
            return;
        }

        createSettingNotification({ payload });
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
        ? (resolvedDialogOptions.entityTitle ?? "Notification Target")
        : "Create or update a Notification Target";
    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const initialValues = notificationTarget
        ? {
              name: notificationTarget.name,
              emailEnabled: notificationTarget.viaEmail?.enabled ?? false,
              senderEmailAccountId: notificationTarget.viaEmail?.sender?.id ?? "",
              notifyAdmins: notificationTarget.viaEmail?.toAllAdmins ?? false,
              notifyProjectOwners: notificationTarget.viaEmail?.toProjectOwners ?? false,
              notifyProjectMembers: notificationTarget.viaEmail?.toProjectMembers ?? false,
              customAddresses: joinAddresses(notificationTarget.viaEmail?.toAddresses),
              slackEnabled: notificationTarget.viaSlack?.enabled ?? false,
              slackWebhookId: notificationTarget.viaSlack?.webhook?.id ?? "",
              discordEnabled: notificationTarget.viaDiscord?.enabled ?? false,
              discordWebhookId: notificationTarget.viaDiscord?.webhook?.id ?? "",
              minSendInterval: notificationTarget.minSendInterval || "3m",
              availableInProjects: notificationTarget.availableInProjects ?? false,
              default: notificationTarget.default ?? false,
          }
        : undefined;
    const isDetailLoading = state.mode === "edit" && detailQuery.isFetching;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[760px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>{dialogTitle}</DialogTitle>
                </DialogHeader>
                {isDetailLoading && <AppLoader />}
                {state.mode !== "closed" && !isDetailLoading && (state.mode === "open" || initialValues) && (
                    <CreateOrEditNotificationTargetForm
                        scope={state.scope}
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
