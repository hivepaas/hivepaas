import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectCloudStorageCommands } from "~/projects/data/commands";
import { ProjectCloudStorageQueries } from "~/projects/data/queries";
import { CloudStorageCommands } from "~/settings/data/commands";
import { CloudStorageQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { UpdateCloudStorageStatusForm } from "../form";
import { useUpdateCloudStorageStatusDialogState } from "../hooks";
import type { UpdateCloudStorageStatusFormOutput } from "../schemas";

export function UpdateCloudStorageStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateCloudStorageStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: updateSettingMeta, isPending: isUpdatingSetting } = CloudStorageCommands.useUpdateMeta({
        onSuccess: () => {
            toast.success("Cloud storage status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateProjectMeta, isPending: isUpdatingProject } = ProjectCloudStorageCommands.useUpdateMeta({
        onSuccess: () => {
            toast.success("Project cloud storage status updated successfully");
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
    const settingDetailQuery = CloudStorageQueries.useFindOneById(
        { id: detailId },
        { enabled: state.mode === "open" && state.scope.type === "settings" },
    );
    const projectDetailQuery = ProjectCloudStorageQueries.useFindOneById(
        {
            projectID: state.mode === "open" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        { enabled: state.mode === "open" && state.scope.type === "project" },
    );
    const detailQuery =
        state.mode === "open" && state.scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const cloudStorage = detailQuery.data?.data;

    function onSubmit(values: UpdateCloudStorageStatusFormOutput) {
        if (state.mode !== "open" || !cloudStorage) return;

        const payload = {
            updateVer: cloudStorage.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            availableInProjects: state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectMeta({ projectID: state.scope.projectId, id: cloudStorage.id, payload });
            return;
        }

        updateSettingMeta({ id: cloudStorage.id, payload });
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
        ? `${resolvedDialogOptions.entityTitle ?? "Cloud Storage"} Status`
        : "Change status";
    const isPending = isUpdatingSetting || isUpdatingProject;
    const showAvailableInProjects = state.mode === "open" && state.scope.type === "settings";
    const initialValues = cloudStorage
        ? {
              status: cloudStorage.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: cloudStorage.expireAt ?? undefined,
              availableInProjects: cloudStorage.availableInProjects ?? false,
              default: cloudStorage.default ?? false,
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
                    <UpdateCloudStorageStatusForm
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
