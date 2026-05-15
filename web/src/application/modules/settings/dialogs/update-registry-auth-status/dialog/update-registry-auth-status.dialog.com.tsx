import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectRegistryAuthCommands } from "~/projects/data/commands";
import { ProjectRegistryAuthQueries } from "~/projects/data/queries";
import { RegistryAuthCommands } from "~/settings/data/commands";
import { RegistryAuthQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { UpdateRegistryAuthStatusForm } from "../form";
import { useUpdateRegistryAuthStatusDialogState } from "../hooks";
import type { UpdateRegistryAuthStatusFormOutput } from "../schemas";

export function UpdateRegistryAuthStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateRegistryAuthStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: updateSettingMeta, isPending: isUpdatingSetting } = RegistryAuthCommands.useUpdateMeta({
        onSuccess: () => {
            toast.success("Registry auth status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
    });

    const { mutate: updateProjectMeta, isPending: isUpdatingProject } = ProjectRegistryAuthCommands.useUpdateMeta({
        onSuccess: () => {
            toast.success("Project registry auth status updated successfully");
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
    const settingDetailQuery = RegistryAuthQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "open" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectRegistryAuthQueries.useFindOneById(
        {
            projectID: state.mode === "open" && state.scope.type === "project" ? state.scope.projectId : "",
            id: detailId,
        },
        {
            enabled: state.mode === "open" && state.scope.type === "project",
        },
    );
    const detailQuery =
        state.mode === "open" && state.scope.type === "project" ? projectDetailQuery : settingDetailQuery;
    const registryAuth = detailQuery.data?.data;

    function onSubmit(values: UpdateRegistryAuthStatusFormOutput) {
        if (state.mode !== "open" || !registryAuth) {
            return;
        }

        const payload = {
            updateVer: registryAuth.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            availableInProjects: state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectMeta({
                projectID: state.scope.projectId,
                id: registryAuth.id,
                payload,
            });
            return;
        }

        updateSettingMeta({
            id: registryAuth.id,
            payload,
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
    const isPending = isUpdatingSetting || isUpdatingProject;
    const showAvailableInProjects = state.mode === "open" && state.scope.type === "settings";
    const initialValues = registryAuth
        ? {
              status: registryAuth.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: registryAuth.expireAt ?? undefined,
              availableInProjects: registryAuth.availableInProjects ?? false,
              default: registryAuth.default ?? false,
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
                    <DialogTitle>Change status</DialogTitle>
                </DialogHeader>
                {isDetailLoading && <AppLoader />}
                {state.mode === "open" && !isDetailLoading && initialValues && (
                    <UpdateRegistryAuthStatusForm
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
