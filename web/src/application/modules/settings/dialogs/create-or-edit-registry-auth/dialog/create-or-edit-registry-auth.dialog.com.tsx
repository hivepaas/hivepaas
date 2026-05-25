import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectRegistryAuthCommands } from "~/projects/data/commands";
import { ProjectRegistryAuthQueries } from "~/projects/data/queries";
import { RegistryAuthCommands } from "~/settings/data/commands";
import { RegistryAuthQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";

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

    const detailId = state.mode === "edit" ? state.id : "";
    const settingDetailQuery = RegistryAuthQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "edit" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectRegistryAuthQueries.useFindOneById(
        {
            projectID: state.mode === "edit" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        {
            enabled: state.mode === "edit" && state.scope.type === "project",
        },
    );
    const detailQuery =
        state.mode === "edit" && state.scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const registryAuth = detailQuery.data?.data;

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

        if (state.mode === "edit" && registryAuth) {
            const updatePayload = {
                ...payload,
                updateVer: registryAuth.updateVer,
            };

            if (state.scope.type === "project") {
                updateProjectRegistryAuth({
                    projectID: state.scope.projectId,
                    id: registryAuth.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingRegistryAuth({
                id: registryAuth.id,
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
        ? (resolvedDialogOptions.entityTitle ?? "Registry Auth")
        : "Create or update a registry auth";
    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const initialValues = registryAuth
        ? {
              name: registryAuth.name,
              address: registryAuth.address,
              username: registryAuth.username,
              readonly: registryAuth.readonly,
              availableInProjects: registryAuth.availableInProjects ?? false,
              default: registryAuth.default ?? false,
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
                    <CreateOrEditRegistryAuthForm
                        isPending={isPending}
                        isTesting={isTesting}
                        testStatus={testStatus}
                        onSubmit={onSubmit}
                        onTestConnection={onTestConnection}
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
