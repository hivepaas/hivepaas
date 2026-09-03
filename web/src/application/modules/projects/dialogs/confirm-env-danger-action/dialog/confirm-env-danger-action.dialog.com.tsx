import { Dialog, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectsCommands } from "~/projects/data/commands";
import { EProjectEnvStatus } from "~/projects/module-shared/enums";

import { useConditionalProject } from "@application/shared/permissions";

import { Separator } from "@/components/ui";

import { ConfirmEnvDangerActionForm } from "../form";
import { useConfirmEnvDangerActionDialogState } from "../hooks";
import type { ConfirmEnvDangerActionFormOutput } from "../schemas";
import { EnvDangerAction } from "../types";

const dialogTitle = {
    [EnvDangerAction.Disable]: "Disable environment",
    [EnvDangerAction.ReEnable]: "Re-enable environment",
    [EnvDangerAction.Delete]: "Delete environment",
} as const;

export function ConfirmEnvDangerActionDialog() {
    const { state, props: dialogOptions, ...actions } = useConfirmEnvDangerActionDialogState();

    const open = state.mode !== "closed";
    const action = state.mode === "open" ? state.action : null;
    const target = state.mode === "open" ? state.target : null;
    const projectPermissions = useConditionalProject({ projectId: target?.projectId ?? "" });

    const { mutate: updateEnvStatus, isPending: isUpdating } = ProjectsCommands.useUpdateEnvStatus({
        onSuccess: (_response, request) => {
            const isReEnable = request.payload.status === EProjectEnvStatus.Active;
            toast.success(isReEnable ? "Environment re-enabled" : "Environment disabled");
            actions.close();
            dialogOptions?.onSuccess?.(isReEnable ? EnvDangerAction.ReEnable : EnvDangerAction.Disable);
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    const { mutate: deleteEnv, isPending: isDeleting } = ProjectsCommands.useDeleteEnv({
        onSuccess: () => {
            toast.success("Environment deleted");
            actions.close();
            dialogOptions?.onSuccess?.(EnvDangerAction.Delete);
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    const isPending = isUpdating || isDeleting;
    const hasRequiredAccess =
        action === EnvDangerAction.Delete ? projectPermissions.canDelete : projectPermissions.canWrite;

    function handleSubmit(_values: ConfirmEnvDangerActionFormOutput) {
        if (!hasRequiredAccess || !target || !action) {
            return;
        }

        if (action === EnvDangerAction.Delete) {
            deleteEnv({
                projectID: target.projectId,
                envName: target.envName,
            });
            return;
        }

        updateEnvStatus({
            projectID: target.projectId,
            envName: target.envName,
            payload: {
                updateVer: target.updateVer,
                status: action === EnvDangerAction.ReEnable ? EProjectEnvStatus.Active : EProjectEnvStatus.Disabled,
            },
        });
    }

    function handleClose(): void {
        if (isPending) {
            return;
        }

        actions.close();
        dialogOptions?.onClose?.();
    }

    return (
        <Dialog
            open={open}
            onOpenChange={nextOpen => {
                if (!nextOpen) {
                    handleClose();
                }
            }}
        >
            <DialogFixedContent className="sm:max-w-[560px]">
                {action && target ? (
                    <>
                        <DialogHeader>
                            <DialogTitle>{dialogTitle[action]}</DialogTitle>
                        </DialogHeader>
                        <div className="px-4">
                            <Separator className="opacity-50" />
                        </div>
                        <ConfirmEnvDangerActionForm
                            action={action}
                            envName={target.envName}
                            isPending={isPending}
                            onSubmit={handleSubmit}
                            readOnly={!hasRequiredAccess}
                        />
                    </>
                ) : null}
            </DialogFixedContent>
        </Dialog>
    );
}
