import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectAccessTokenCommands } from "~/projects/data/commands";
import { ProjectAccessTokenQueries } from "~/projects/data/queries";
import { AccessTokenCommands } from "~/settings/data/commands";
import { AccessTokenQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { UpdateAccessTokenStatusForm } from "../form";
import { useUpdateAccessTokenStatusDialogState } from "../hooks";
import type { UpdateAccessTokenStatusFormOutput } from "../schemas";

export function UpdateAccessTokenStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateAccessTokenStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: updateSettingMeta, isPending: isUpdatingSetting } = AccessTokenCommands.useUpdateMeta({
        onSuccess: () => {
            toast.success("Access token status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateProjectMeta, isPending: isUpdatingProject } = ProjectAccessTokenCommands.useUpdateMeta({
        onSuccess: () => {
            toast.success("Project access token status updated successfully");
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
    const settingDetailQuery = AccessTokenQueries.useFindOneById(
        { id: detailId },
        { enabled: state.mode === "open" && state.scope.type === "settings" },
    );
    const projectDetailQuery = ProjectAccessTokenQueries.useFindOneById(
        {
            projectID: state.mode === "open" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        { enabled: state.mode === "open" && state.scope.type === "project" },
    );
    const detailQuery =
        state.mode === "open" && state.scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const accessToken = detailQuery.data?.data;

    function onSubmit(values: UpdateAccessTokenStatusFormOutput) {
        if (state.mode !== "open" || !accessToken) return;

        const payload = {
            updateVer: accessToken.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            availableInProjects: state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectMeta({ projectID: state.scope.projectId, id: accessToken.id, payload });
            return;
        }

        updateSettingMeta({ id: accessToken.id, payload });
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
        ? `${resolvedDialogOptions.entityTitle ?? "Access Token"} Status`
        : "Change status";
    const isPending = isUpdatingSetting || isUpdatingProject;
    const showAvailableInProjects = state.mode === "open" && state.scope.type === "settings";
    const initialValues = accessToken
        ? {
              status: accessToken.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: accessToken.expireAt ?? undefined,
              availableInProjects: accessToken.availableInProjects ?? false,
              default: accessToken.default ?? false,
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
                    <UpdateAccessTokenStatusForm
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
