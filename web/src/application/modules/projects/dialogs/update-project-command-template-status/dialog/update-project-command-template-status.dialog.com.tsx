import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectCommandTemplateCommands } from "~/projects/data/commands";
import { ProjectCommandTemplateQueries } from "~/projects/data/queries";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { UpdateProjectCommandTemplateStatusForm } from "../form";
import { useUpdateProjectCommandTemplateStatusDialogState } from "../hooks";
import type { UpdateProjectCommandTemplateStatusFormOutput } from "../schemas";

export function UpdateProjectCommandTemplateStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateProjectCommandTemplateStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useSettingsScopePermissions({ type: "project" });

    const { mutate: updateProjectStatus, isPending } = ProjectCommandTemplateCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Project Command Template status updated successfully");
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
    const detailQuery = ProjectCommandTemplateQueries.useFindOneById(
        {
            projectID: state.mode === "open" ? state.projectId : "",
            id: detailId,
        },
        {
            enabled: state.mode === "open",
        },
    );
    const commandTemplate = detailQuery.data?.data;

    function onSubmit(values: UpdateProjectCommandTemplateStatusFormOutput) {
        if (state.mode !== "open" || !commandTemplate) {
            return;
        }

        updateProjectStatus({
            projectID: state.projectId,
            id: commandTemplate.id,
            payload: {
                updateVer: commandTemplate.updateVer,
                status: values.status,
                expireAt: values.expireAt ?? null,
                availableInProjects: false,
                default: values.default,
            },
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
    const readOnlyInherited = dialogOptions?.readOnlyInherited === true;
    const dialogTitle = readOnlyInherited ? "Command Template Status" : "Change status";
    const isDetailLoading = state.mode === "open" && detailQuery.isFetching;
    const initialValues = commandTemplate
        ? {
              status:
                  commandTemplate.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: commandTemplate.expireAt ?? undefined,
              default: commandTemplate.default,
          }
        : undefined;

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
                    <UpdateProjectCommandTemplateStatusForm
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                        readOnlyInherited={readOnlyInherited}
                        readOnly={!canWrite}
                        onClose={handleClose}
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
