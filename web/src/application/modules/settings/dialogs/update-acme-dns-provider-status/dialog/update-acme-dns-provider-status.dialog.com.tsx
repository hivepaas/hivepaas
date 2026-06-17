import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectAcmeDnsProviderCommands } from "~/projects/data/commands";
import { ProjectAcmeDnsProviderQueries } from "~/projects/data/queries";
import { AcmeDnsProviderCommands } from "~/settings/data/commands";
import { AcmeDnsProviderQueries } from "~/settings/data/queries";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { UpdateAcmeDnsProviderStatusForm } from "../form";
import { useUpdateAcmeDnsProviderStatusDialogState } from "../hooks";
import type { UpdateAcmeDnsProviderStatusFormOutput } from "../schemas";

export function UpdateAcmeDnsProviderStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateAcmeDnsProviderStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const permissionScope = state.mode === "closed" ? ({ type: "settings" } as const) : state.scope;
    const { canWrite } = useSettingsScopePermissions(permissionScope);

    const { mutate: updateSettingStatus, isPending: isUpdatingSetting } = AcmeDnsProviderCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("ACME DNS provider status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });

    const { mutate: updateProjectStatus, isPending: isUpdatingProject } =
        ProjectAcmeDnsProviderCommands.useUpdateStatus({
            onSuccess: () => {
                toast.success("Project ACME DNS provider status updated successfully");
                closeDialog();
                dialogOptions?.onSuccess?.();
            },
            onError: dialogOptions?.onError,
        });

    useEffect(() => {
        if (state.mode === "closed") {
            setHasChanges(false);
            clearDialog();
        }
    }, [clearDialog, state.mode]);

    const detailId = state.mode === "open" ? state.id : "";
    const settingDetailQuery = AcmeDnsProviderQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "open" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectAcmeDnsProviderQueries.useFindOneById(
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
    const acmeDnsProvider = detailQuery.data?.data;

    function onSubmit(values: UpdateAcmeDnsProviderStatusFormOutput) {
        if (state.mode !== "open" || !acmeDnsProvider) {
            return;
        }

        const payload = {
            updateVer: acmeDnsProvider.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            availableInProjects: state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectStatus({
                projectID: state.scope.projectId,
                id: acmeDnsProvider.id,
                payload,
            });
            return;
        }

        updateSettingStatus({
            id: acmeDnsProvider.id,
            payload,
        });
    }

    function handleClose() {
        if (isPending) {
            return;
        }

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
        ? `${resolvedDialogOptions.entityTitle ?? "ACME DNS Provider"} Status`
        : "Change status";
    const isPending = isUpdatingSetting || isUpdatingProject;
    const showAvailableInProjects = state.mode === "open" && state.scope.type === "settings";
    const initialValues = acmeDnsProvider
        ? {
              status:
                  acmeDnsProvider.status === ESettingStatus.Disabled
                      ? ESettingStatus.Disabled
                      : ESettingStatus.Active,
              expireAt: acmeDnsProvider.expireAt ?? undefined,
              availableInProjects: acmeDnsProvider.availableInProjects ?? false,
              default: acmeDnsProvider.default ?? false,
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
                    <UpdateAcmeDnsProviderStatusForm
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
