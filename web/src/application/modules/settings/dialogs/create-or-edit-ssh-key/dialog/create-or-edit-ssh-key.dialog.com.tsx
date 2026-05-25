import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectSSHKeyCommands } from "~/projects/data/commands";
import { ProjectSSHKeyQueries } from "~/projects/data/queries";
import { SSHKeyCommands } from "~/settings/data/commands";
import { SSHKeyQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";

import { CreateOrEditSSHKeyForm } from "../form";
import { useCreateOrEditSSHKeyDialogState } from "../hooks";
import type { CreateOrEditSSHKeyFormOutput } from "../schemas";

export function CreateOrEditSSHKeyDialog() {
    const { state, props: dialogOptions, close: closeDialog, clear: clearDialog } = useCreateOrEditSSHKeyDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: createSettingSSHKey, isPending: isCreatingSetting } = SSHKeyCommands.useCreateOne({
        onSuccess: () => {
            toast.success("SSH key created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateSettingSSHKey, isPending: isUpdatingSetting } = SSHKeyCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("SSH key updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: createProjectSSHKey, isPending: isCreatingProject } = ProjectSSHKeyCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project SSH key created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateProjectSSHKey, isPending: isUpdatingProject } = ProjectSSHKeyCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project SSH key updated successfully");
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
    const settingDetailQuery = SSHKeyQueries.useFindOneById(
        { id: detailId },
        { enabled: state.mode === "edit" && state.scope.type === "settings" },
    );
    const projectDetailQuery = ProjectSSHKeyQueries.useFindOneById(
        {
            projectID: state.mode === "edit" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        { enabled: state.mode === "edit" && state.scope.type === "project" },
    );
    const detailQuery =
        state.mode === "edit" && state.scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const sshKey = detailQuery.data?.data;

    function createPayload(values: CreateOrEditSSHKeyFormOutput) {
        return {
            availableInProjects:
                state.mode !== "closed" && state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            name: values.name,
            privateKey: values.privateKey,
            passphrase: values.passphrase,
            targets: values.targets.map(item => item.value),
        };
    }

    function onSubmit(values: CreateOrEditSSHKeyFormOutput) {
        if (state.mode === "closed") return;

        const payload = createPayload(values);

        if (state.mode === "edit" && sshKey) {
            const updatePayload = { ...payload, updateVer: sshKey.updateVer };

            if (state.scope.type === "project") {
                updateProjectSSHKey({ projectID: state.scope.projectId, id: sshKey.id, payload: updatePayload });
                return;
            }

            updateSettingSSHKey({ id: sshKey.id, payload: updatePayload });
            return;
        }

        if (state.scope.type === "project") {
            createProjectSSHKey({ projectID: state.scope.projectId, payload });
            return;
        }

        createSettingSSHKey({ payload });
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
        ? (resolvedDialogOptions.entityTitle ?? "SSH Key")
        : "Create or update an SSH key";
    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const initialValues = sshKey
        ? {
              name: sshKey.name,
              privateKey: sshKey.privateKey,
              passphrase: sshKey.passphrase ?? "",
              targets: (sshKey.targets ?? []).map(value => ({ value })),
              availableInProjects: sshKey.availableInProjects ?? false,
              default: sshKey.default ?? false,
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
                    <CreateOrEditSSHKeyForm
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
