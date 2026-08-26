import { useEffect, useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectCommandPipeCommands } from "~/projects/data/commands";
import { ProjectCommandPipeQueries } from "~/projects/data/queries";
import { useSettingsScopePermissions } from "~/settings/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { UpdateProjectCommandPipeStatusForm } from "../form";
import { useUpdateProjectCommandPipeStatusDialogState } from "../hooks";
import type { UpdateProjectCommandPipeStatusFormOutput } from "../schemas";

export function UpdateProjectCommandPipeStatusDialog() {
    const {
        state,
        props: dialogOptions,
        close: closeDialog,
        clear: clearDialog,
    } = useUpdateProjectCommandPipeStatusDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useSettingsScopePermissions({ type: "project" });

    const { mutate: updateProjectStatus, isPending } = ProjectCommandPipeCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success("Project Command Pipe status updated successfully");
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
    const detailQuery = ProjectCommandPipeQueries.useFindOneById(
        {
            projectID: state.mode === "open" ? state.projectId : "",
            env: state.mode === "open" ? state.env : undefined,
            id: detailId,
        },
        {
            enabled: state.mode === "open",
        },
    );
    const commandPipe = detailQuery.data?.data;

    function onSubmit(values: UpdateProjectCommandPipeStatusFormOutput) {
        if (state.mode !== "open" || !commandPipe) {
            return;
        }

        updateProjectStatus({
            projectID: state.projectId,
            env: state.env,
            id: commandPipe.id,
            payload: {
                updateVer: commandPipe.updateVer,
                status: values.status,
                expireAt: values.expireAt ?? null,
                inheritable: false,
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
    const dialogTitle = readOnlyInherited ? "Command Pipe Status" : "Change status";
    const isDetailLoading = state.mode === "open" && detailQuery.isFetching;
    const initialValues = commandPipe
        ? {
              status: commandPipe.status === ESettingStatus.Disabled ? ESettingStatus.Disabled : ESettingStatus.Active,
              expireAt: commandPipe.expireAt ?? undefined,
              default: commandPipe.default,
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
                    <UpdateProjectCommandPipeStatusForm
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
