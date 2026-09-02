import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectBackupRepoCommands } from "~/projects/data/commands";
import { ProjectBackupRepoQueries } from "~/projects/data/queries";
import { BackupRepoCommands } from "~/settings/data/commands";
import { BackupRepoQueries } from "~/settings/data/queries";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { UpdateBackupRepoStatusForm } from "../form";
import { useUpdateBackupRepoStatusDialogState } from "../hooks";
import type { UpdateBackupRepoStatusFormOutput } from "../schemas";

export function UpdateBackupRepoStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateBackupRepoStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const permissionScope = state.mode === "closed" ? ({ type: "settings" } as const) : state.scope;
    const { canWrite } = useSettingsScopePermissions(permissionScope);

    const { mutate: updateSettingStatus, isPending: isUpdatingSetting } = BackupRepoCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Backup repository status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateProjectStatus, isPending: isUpdatingProject } = ProjectBackupRepoCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Project backup repository status updated successfully");
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
    const settingDetailQuery = BackupRepoQueries.useFindOneById(
        { id: detailId },
        { enabled: state.mode === "open" && state.scope.type === "settings" },
    );
    const projectDetailQuery = ProjectBackupRepoQueries.useFindOneById(
        {
            projectID: state.mode === "open" && state.scope.type === "project" ? state.scope.projectId : "",
            env: state.mode === "open" && state.scope.type === "project" ? state.scope.env : undefined,
            id: detailId,
        },
        { enabled: state.mode === "open" && state.scope.type === "project" },
    );
    const detailQuery =
        state.mode === "open" && state.scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const backupRepo = detailQuery.data?.data;

    function onSubmit(values: UpdateBackupRepoStatusFormOutput) {
        if (state.mode !== "open" || !backupRepo) return;

        const payload = {
            updateVer: backupRepo.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            inheritable: state.scope.type === "project" ? false : values.inheritable,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectStatus({
                projectID: state.scope.projectId,
                env: state.scope.env,
                id: backupRepo.id,
                payload,
            });
            return;
        }

        updateSettingStatus({ id: backupRepo.id, payload });
    }

    function handleClose() {
        if (isPending) return;

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
        ? `${resolvedDialogOptions.entityTitle ?? "Backup Repository"} Status`
        : "Change status";
    const isPending = isUpdatingSetting || isUpdatingProject;
    const showAvailableInProjects = state.mode === "open" && state.scope.type === "settings";
    const initialValues = backupRepo
        ? {
              status: backupRepo.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: backupRepo.expireAt ? new Date(backupRepo.expireAt) : undefined,
              inheritable: backupRepo.inheritable ?? false,
              default: backupRepo.default ?? false,
          }
        : undefined;
    const isDetailLoading = state.mode === "open" && detailQuery.isFetching;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>{dialogTitle}</DialogTitle>
                </DialogHeader>
                {isDetailLoading && (
                    <DialogBody>
                        <AppLoader />
                    </DialogBody>
                )}
                {state.mode === "open" && !isDetailLoading && initialValues && (
                    <UpdateBackupRepoStatusForm
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
            </DialogFixedContent>
        </Dialog>
    );
}
