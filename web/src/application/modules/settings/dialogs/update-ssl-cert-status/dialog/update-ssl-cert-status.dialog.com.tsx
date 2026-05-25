import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectSslCertCommands } from "~/projects/data/commands";
import { ProjectSslCertQueries } from "~/projects/data/queries";
import { SslCertCommands } from "~/settings/data/commands";
import { SslCertQueries } from "~/settings/data/queries";

import { AppLoader } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { UpdateSslCertStatusForm } from "../form";
import { useUpdateSslCertStatusDialogState } from "../hooks";
import type { UpdateSslCertStatusFormOutput } from "../schemas";

export function UpdateSslCertStatusDialog() {
    const { state, props: dialogOptions, close: closeDialog, clear: clearDialog } = useUpdateSslCertStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: updateSettingStatus, isPending: isUpdatingSetting } = SslCertCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("SSL certificate status updated successfully");
            closeDialog();
            dialogOptions?.onSuccess?.();
        },
        onError: dialogOptions?.onError,
    });

    const { mutate: updateProjectStatus, isPending: isUpdatingProject } = ProjectSslCertCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Project SSL certificate status updated successfully");
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
    const settingDetailQuery = SslCertQueries.useFindOneById(
        { id: detailId },
        {
            enabled: state.mode === "open" && state.scope.type === "settings",
        },
    );
    const projectDetailQuery = ProjectSslCertQueries.useFindOneById(
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
    const sslCert = detailQuery.data?.data;

    function onSubmit(values: UpdateSslCertStatusFormOutput) {
        if (state.mode !== "open" || !sslCert) {
            return;
        }

        const payload = {
            updateVer: sslCert.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            availableInProjects: state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectStatus({
                projectID: state.scope.projectId,
                id: sslCert.id,
                payload,
            });
            return;
        }

        updateSettingStatus({
            id: sslCert.id,
            payload,
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
        ? `${resolvedDialogOptions.entityTitle ?? "SSL Certificate"} Status`
        : "Change status";
    const isPending = isUpdatingSetting || isUpdatingProject;
    const showAvailableInProjects = state.mode === "open" && state.scope.type === "settings";
    const initialValues = sslCert
        ? {
              status: sslCert.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: sslCert.expireAt ?? undefined,
              availableInProjects: sslCert.availableInProjects ?? false,
              default: sslCert.default ?? false,
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
                    <UpdateSslCertStatusForm
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
