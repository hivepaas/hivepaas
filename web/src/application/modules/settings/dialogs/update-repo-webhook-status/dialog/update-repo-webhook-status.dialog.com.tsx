import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectRepoWebhookCommands } from "~/projects/data/commands";
import { ProjectRepoWebhookQueries } from "~/projects/data/queries";
import { RepoWebhookCommands } from "~/settings/data/commands";
import { RepoWebhookQueries } from "~/settings/data/queries";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { UpdateRepoWebhookStatusForm } from "../form";
import { useUpdateRepoWebhookStatusDialogState } from "../hooks";
import type { UpdateRepoWebhookStatusFormOutput } from "../schemas";

export function UpdateRepoWebhookStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateRepoWebhookStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const permissionScope = state.mode === "closed" ? ({ type: "settings" } as const) : state.scope;
    const { canWrite } = useSettingsScopePermissions(permissionScope);

    const { mutate: updateSettingsStatus, isPending: isUpdatingSettings } = RepoWebhookCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Webhook status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });

    const { mutate: updateProjectStatus, isPending: isUpdatingProject } = ProjectRepoWebhookCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Project webhook status updated successfully");
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
    const settingsDetailQuery = RepoWebhookQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "open" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectRepoWebhookQueries.useFindOneById(
        {
            projectID: state.mode === "open" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        {
            enabled: state.mode === "open" && state.scope.type === "project",
        },
    );
    const detailQuery =
        state.mode === "open" && state.scope.type === "project" ? projectDetailQuery : settingsDetailQuery;
    const repoWebhook = detailQuery.data?.data;

    function onSubmit(values: UpdateRepoWebhookStatusFormOutput) {
        if (state.mode !== "open" || !repoWebhook) {
            return;
        }

        const payload = {
            updateVer: repoWebhook.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            availableInProjects: state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectStatus({
                projectID: state.scope.projectId,
                id: repoWebhook.id,
                payload,
            });
            return;
        }

        updateSettingsStatus({
            id: repoWebhook.id,
            payload,
        });
    }

    function handleClose() {
        if (isPending) {
            return;
        }

        if (
            !readOnlyInherited &&
            canWrite &&
            hasChanges &&
            !window.confirm("Are you sure you want to close without saving changes?")
        ) {
            return;
        }

        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const resolvedDialogOptions = dialogOptions ?? {};
    const readOnlyInherited = resolvedDialogOptions.readOnlyInherited === true;
    const dialogTitle = readOnlyInherited
        ? `${resolvedDialogOptions.entityTitle ?? "Webhook"} Status`
        : "Change status";
    const isPending = isUpdatingSettings || isUpdatingProject;
    const showAvailableInProjects = state.mode === "open" && state.scope.type === "settings";
    const initialValues = repoWebhook
        ? {
              status: repoWebhook.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: repoWebhook.expireAt ?? undefined,
              availableInProjects: repoWebhook.availableInProjects ?? false,
              default: repoWebhook.default ?? false,
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
                    <UpdateRepoWebhookStatusForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        showAvailableInProjects={showAvailableInProjects}
                        readOnlyInherited={readOnlyInherited}
                        readOnly={!canWrite}
                        onClose={handleClose}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
