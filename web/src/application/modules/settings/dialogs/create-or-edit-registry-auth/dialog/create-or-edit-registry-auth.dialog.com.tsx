import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectRegistryAuthCommands } from "~/projects/data/commands";
import { RegistryAuthCommands } from "~/settings/data/commands";

import { CreateOrEditRegistryAuthForm } from "../form";
import { useCreateOrEditRegistryAuthDialogState } from "../hooks";
import type { CreateOrEditRegistryAuthFormOutput } from "../schemas";

export function CreateOrEditRegistryAuthDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useCreateOrEditRegistryAuthDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const [testStatus, setTestStatus] = useState<"idle" | "succeeded" | "failed">("idle");

    const { mutate: createSettingRegistryAuth, isPending: isCreatingSetting } = RegistryAuthCommands.useCreateOne({
        onSuccess: () => {
            toast.success("Registry auth created successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: updateSettingRegistryAuth, isPending: isUpdatingSetting } = RegistryAuthCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Registry auth updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });
    const { mutate: createProjectRegistryAuth, isPending: isCreatingProject } =
        ProjectRegistryAuthCommands.useCreateOne({
            onSuccess: () => {
                toast.success("Project registry auth created successfully");
                closeDialog();
                dialogOptions?.onSuccess?.();
            },
        });
    const { mutate: updateProjectRegistryAuth, isPending: isUpdatingProject } =
        ProjectRegistryAuthCommands.useUpdateOne({
            onSuccess: () => {
                toast.success("Project registry auth updated successfully");
                closeDialog();
                dialogOptions?.onSuccess?.();
            },
        });
    const { mutate: testConnection, isPending: isTesting } = RegistryAuthCommands.useTestConn({
        onSuccess: () => {
            setTestStatus("succeeded");
        },
        onError: () => {
            setTestStatus("failed");
        },
    });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            setTestStatus("idle");
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    function createPayload(values: CreateOrEditRegistryAuthFormOutput) {
        return {
            availableInProjects:
                state.mode !== "closed" && state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
            name: values.name,
            address: values.address,
            username: values.username,
            password: values.password,
            readonly: values.readonly,
        };
    }

    function onSubmit(values: CreateOrEditRegistryAuthFormOutput) {
        if (state.mode === "closed") {
            return;
        }

        const payload = createPayload(values);

        if (state.mode === "edit") {
            const updatePayload = {
                ...payload,
                updateVer: state.registryAuth.updateVer,
            };

            if (state.scope.type === "project") {
                updateProjectRegistryAuth({
                    projectID: state.scope.projectId,
                    id: state.registryAuth.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingRegistryAuth({
                id: state.registryAuth.id,
                payload: updatePayload,
            });
            return;
        }

        if (state.scope.type === "project") {
            createProjectRegistryAuth({
                projectID: state.scope.projectId,
                payload,
            });
            return;
        }

        createSettingRegistryAuth({ payload });
    }

    function onTestConnection(values: CreateOrEditRegistryAuthFormOutput) {
        setTestStatus("idle");
        testConnection({
            payload: {
                name: values.name,
                address: values.address,
                username: values.username,
                password: values.password,
                readonly: values.readonly,
            },
        });
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
                  name: state.registryAuth.name,
                  address: state.registryAuth.address,
                  username: state.registryAuth.username,
                  readonly: state.registryAuth.readonly,
                  availableInProjects: state.registryAuth.availableInProjects ?? false,
                  default: state.registryAuth.default ?? false,
              }
            : undefined;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[390px] w-[760px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Create or update a registry auth</DialogTitle>
                </DialogHeader>
                {state.mode !== "closed" && (
                    <CreateOrEditRegistryAuthForm
                        isPending={isPending}
                        isTesting={isTesting}
                        testStatus={testStatus}
                        onSubmit={onSubmit}
                        onTestConnection={onTestConnection}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        showAvailableInProjects={showAvailableInProjects}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
