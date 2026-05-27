import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectGithubAppCommands } from "~/projects/data/commands";
import { ProjectGithubAppQueries } from "~/projects/data/queries";
import { GithubAppCommands } from "~/settings/data/commands";
import { GithubAppQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { UpdateGithubAppStatusForm } from "../form";
import { useUpdateGithubAppStatusDialogState } from "../hooks";
import type { UpdateGithubAppStatusFormOutput } from "../schemas";

export function UpdateGithubAppStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateGithubAppStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: updateSettingsStatus, isPending: isUpdatingSettings } = GithubAppCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Github app status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });

    const { mutate: updateProjectStatus, isPending: isUpdatingProject } = ProjectGithubAppCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Project github app status updated successfully");
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
    const settingsDetailQuery = GithubAppQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "open" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectGithubAppQueries.useFindOneById(
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
    const githubApp = detailQuery.data?.data;

    function onSubmit(values: UpdateGithubAppStatusFormOutput) {
        if (state.mode !== "open" || !githubApp) {
            return;
        }

        const payload = {
            updateVer: githubApp.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            availableInProjects: state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectStatus({
                projectID: state.scope.projectId,
                id: githubApp.id,
                payload,
            });
            return;
        }

        updateSettingsStatus({
            id: githubApp.id,
            payload,
        });
    }

    function handleClose() {
        if (isPending) {
            return;
        }

        if (
            !readOnlyInherited &&
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
        ? `${resolvedDialogOptions.entityTitle ?? "Github App"} Status`
        : "Change status";
    const isPending = isUpdatingSettings || isUpdatingProject;
    const showAvailableInProjects = state.mode === "open" && state.scope.type === "settings";
    const initialValues = githubApp
        ? {
              status: githubApp.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: githubApp.expireAt ?? undefined,
              availableInProjects: githubApp.availableInProjects ?? false,
              default: githubApp.default ?? false,
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
                    <UpdateGithubAppStatusForm
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
