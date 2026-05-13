import { useEffect, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectSslCertCommands } from "~/projects/data/commands";
import { SslCertCommands } from "~/settings/data/commands";

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

    function onSubmit(values: UpdateSslCertStatusFormOutput) {
        if (state.mode !== "open") {
            return;
        }

        const payload = {
            updateVer: state.sslCert.updateVer,
            status: values.status,
            expireAt: values.expireAt ?? null,
            availableInProjects: state.scope.type === "project" ? false : values.availableInProjects,
            default: values.default,
        };

        if (state.scope.type === "project") {
            updateProjectStatus({
                projectID: state.scope.projectId,
                id: state.sslCert.id,
                payload,
            });
            return;
        }

        updateSettingStatus({
            id: state.sslCert.id,
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
    const initialValues =
        state.mode === "open"
            ? {
                  status:
                      state.sslCert.status === ESettingStatus.Disabled
                          ? ESettingStatus.Disabled
                          : ESettingStatus.Active,
                  expireAt: state.sslCert.expireAt ?? undefined,
                  availableInProjects: state.sslCert.availableInProjects ?? false,
                  default: state.sslCert.default ?? false,
              }
            : undefined;

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>Change status</DialogTitle>
                </DialogHeader>
                {state.mode === "open" && (
                    <UpdateSslCertStatusForm
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
