import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectBackupRepoCommands } from "~/projects/data/commands";
import { ProjectBackupRepoQueries } from "~/projects/data/queries";
import { BackupRepoCommands } from "~/settings/data/commands";
import { BackupRepoQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { ChangePasswordForm } from "@application/shared/dialogs/change-password/form";
import type { AccountPasswordFormSchemaOutput } from "@application/shared/dialogs/change-password/schemas";

import { Separator } from "@/components/ui";

import { useUpdateBackupRepoPasswordDialogState } from "../hooks";

export function UpdateBackupRepoPasswordDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateBackupRepoPasswordDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: updateSettingPassword, isPending: isUpdatingSetting } = BackupRepoCommands.useUpdatePassword({
        onSuccess: () => {
            toast.success("Backup repository password updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateProjectPassword, isPending: isUpdatingProject } = ProjectBackupRepoCommands.useUpdatePassword(
        {
            onSuccess: () => {
                toast.success("Project backup repository password updated successfully");
                closeDialog();
                dialogOptions?.onSuccess?.();
            },
        },
    );

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

    const isPending = isUpdatingSetting || isUpdatingProject;
    const isDetailLoading = state.mode === "open" && detailQuery.isFetching && !backupRepo;

    function onSubmit(values: AccountPasswordFormSchemaOutput) {
        if (state.mode !== "open") return;

        const updateVer = backupRepo?.updateVer ?? state.updateVer ?? 0;
        const payload = {
            updateVer,
            currentPassword: values.currentPassword,
            newPassword: values.newPassword,
            inheritable: state.scope.type === "project" ? false : backupRepo?.inheritable,
            default: backupRepo?.default,
        };

        if (state.scope.type === "project") {
            updateProjectPassword({
                projectID: state.scope.projectId,
                env: state.scope.env,
                id: state.id,
                payload,
            });
            return;
        }

        updateSettingPassword({ id: state.id, payload });
    }

    function handleClose() {
        if (isPending) return;

        if (hasChanges && !window.confirm("Are you sure you want to close modal without saving changes?")) {
            return;
        }

        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="min-w-[400px] w-fit">
                <DialogHeader>
                    <DialogTitle>Change Password</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                {isDetailLoading && (
                    <DialogBody>
                        <AppLoader />
                    </DialogBody>
                )}
                {state.mode === "open" && !isDetailLoading && (
                    <ChangePasswordForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
