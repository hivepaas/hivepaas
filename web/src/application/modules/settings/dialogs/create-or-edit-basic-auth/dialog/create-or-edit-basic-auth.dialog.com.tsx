import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectBasicAuthCommands } from "~/projects/data/commands";
import { BasicAuthCommands } from "~/settings/data/commands";

import { CreateOrEditBasicAuthForm } from "../form";
import { useCreateOrEditBasicAuthDialogState } from "../hooks";
import type { CreateOrEditBasicAuthFormOutput } from "../schemas";

export function CreateOrEditBasicAuthDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useCreateOrEditBasicAuthDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: createSettingBasicAuth, isPending: isCreatingSetting } = BasicAuthCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Basic auth created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateSettingBasicAuth, isPending: isUpdatingSetting } = BasicAuthCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Basic auth updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: createProjectBasicAuth, isPending: isCreatingProject } = ProjectBasicAuthCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Project basic auth created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateProjectBasicAuth, isPending: isUpdatingProject } = ProjectBasicAuthCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Project basic auth updated successfully");
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

    function onSubmit(values: CreateOrEditBasicAuthFormOutput) {
        if (state.mode === "closed") {
            return;
        }

        const availableInProjects = state.scope.type === "project" ? false : values.availableInProjects;
        const payload = {
            availableInProjects,
            default: values.default,
            name: values.name,
            username: values.username,
            password: values.password,
        };

        if (state.mode === "edit") {
            const updatePayload = {
                ...payload,
                updateVer: state.basicAuth.updateVer,
            };

            if (state.scope.type === "project") {
                updateProjectBasicAuth({
                    projectID: state.scope.projectId,
                    id: state.basicAuth.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingBasicAuth({
                id: state.basicAuth.id,
                payload: updatePayload,
            });
            return;
        }

        if (state.scope.type === "project") {
            createProjectBasicAuth({
                projectID: state.scope.projectId,
                payload,
            });
            return;
        }

        createSettingBasicAuth({ payload });
    }

    function handleClose() {
        if (isPending) {
            return;
        }

        if (hasChanges && !window.confirm("Are you sure you want to close without saving changes?")) {
            return;
        }

        closeDialog();
        dialogOptions?.onClose?.();
    }

    const open = state.mode !== "closed";
    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const initialValues =
        state.mode === "edit"
            ? {
                  name: state.basicAuth.name,
                  username: state.basicAuth.username,
                  availableInProjects: state.basicAuth.availableInProjects ?? false,
                  default: state.basicAuth.default ?? false,
              }
            : undefined;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[760px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Create or update a basic auth</DialogTitle>
                </DialogHeader>
                {state.mode !== "closed" && (
                    <CreateOrEditBasicAuthForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        showAvailableInProjects={showAvailableInProjects}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
