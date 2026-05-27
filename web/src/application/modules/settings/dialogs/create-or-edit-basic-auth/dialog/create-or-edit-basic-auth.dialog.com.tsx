import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectBasicAuthCommands } from "~/projects/data/commands";
import { ProjectBasicAuthQueries } from "~/projects/data/queries";
import { BasicAuthCommands } from "~/settings/data/commands";
import { BasicAuthQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";

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

    const detailId = state.mode === "edit" ? state.id : "";
    const settingDetailQuery = BasicAuthQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "edit" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectBasicAuthQueries.useFindOneById(
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
    const basicAuth = detailQuery.data?.data;

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

        if (state.mode === "edit" && basicAuth) {
            const updatePayload = {
                ...payload,
                updateVer: basicAuth.updateVer,
            };

            if (state.scope.type === "project") {
                updateProjectBasicAuth({
                    projectID: state.scope.projectId,
                    id: basicAuth.id,
                    payload: updatePayload,
                });
                return;
            }

            updateSettingBasicAuth({
                id: basicAuth.id,
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
        ? (resolvedDialogOptions.entityTitle ?? "Basic Auth")
        : "Create or update a basic auth";
    const isPending = isCreatingSetting || isUpdatingSetting || isCreatingProject || isUpdatingProject;
    const showAvailableInProjects = state.mode !== "closed" && state.scope.type === "settings";
    const initialValues = basicAuth
        ? {
              name: basicAuth.name,
              username: basicAuth.username,
              password: basicAuth.password,
              availableInProjects: basicAuth.availableInProjects ?? false,
              default: basicAuth.default ?? false,
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
                    <CreateOrEditBasicAuthForm
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
