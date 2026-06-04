import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { ProjectsCommands } from "~/projects/data/commands";
import { QK } from "~/projects/data/constants";
import { EProjectStatus } from "~/projects/module-shared/enums";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalProject } from "@application/shared/permissions";

import { ConfirmProjectDangerActionForm } from "../form";
import { useConfirmProjectDangerActionDialogState } from "../hooks";
import type { ConfirmProjectDangerActionFormOutput } from "../schemas";
import { ProjectDangerAction } from "../types";

const dialogTitle = {
    [ProjectDangerAction.Disable]: "Disable project",
    [ProjectDangerAction.ReEnable]: "Re-enable project",
    [ProjectDangerAction.Delete]: "Delete project",
} as const;

export function ConfirmProjectDangerActionDialog() {
    const { state, props: dialogOptions, ...actions } = useConfirmProjectDangerActionDialogState();
    const { navigate } = useAppNavigate();
    const queryClient = useQueryClient();

    const open = state.mode !== "closed";
    const action = state.mode === "open" ? state.action : null;
    const target = state.mode === "open" ? state.target : null;
    const projectPermissions = useConditionalProject({ projectId: target?.projectId ?? "" });

    const { mutate: updateProjectStatus, isPending: isUpdating } = ProjectsCommands.useUpdateStatus({
        onSuccess: (_response, request) => {
            const nextAction =
                request.payload.status === EProjectStatus.Active
                    ? ProjectDangerAction.ReEnable
                    : ProjectDangerAction.Disable;

            toast.success(nextAction === ProjectDangerAction.ReEnable ? "Project re-enabled" : "Project disabled");
            actions.close();
            dialogOptions?.onSuccess?.(nextAction);
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    const { mutate: deleteProject, isPending: isDeleting } = ProjectsCommands.useDeleteOne({
        onSuccess: (_response, request) => {
            toast.success("Project deleted");
            actions.close();
            dialogOptions?.onSuccess?.(ProjectDangerAction.Delete);
            navigate.modules(ROUTE.projects.list.$route, { ignorePrevPath: true });
            window.setTimeout(() => {
                queryClient.removeQueries({
                    queryKey: [QK["projects.$.find-one-by-id"], { projectID: request.projectID }],
                    exact: true,
                });
            }, 0);
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    const isPending = isUpdating || isDeleting;
    const hasRequiredAccess =
        action === ProjectDangerAction.Delete ? projectPermissions.canDelete : projectPermissions.canWrite;

    function handleSubmit(_values: ConfirmProjectDangerActionFormOutput) {
        if (!hasRequiredAccess || !target || !action) {
            return;
        }

        if (action === ProjectDangerAction.Delete) {
            deleteProject({
                projectID: target.projectId,
            });
            return;
        }

        updateProjectStatus({
            projectID: target.projectId,
            payload: {
                updateVer: target.updateVer,
                status: action === ProjectDangerAction.ReEnable ? EProjectStatus.Active : EProjectStatus.Disabled,
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
            <DialogContent className="sm:max-w-[560px]">
                {action && target ? (
                    <>
                        <DialogHeader>
                            <DialogTitle className="text-2xl">{dialogTitle[action]}</DialogTitle>
                        </DialogHeader>
                        <ConfirmProjectDangerActionForm
                            action={action}
                            projectName={target.projectName}
                            isPending={isPending}
                            onSubmit={handleSubmit}
                            readOnly={!hasRequiredAccess}
                        />
                    </>
                ) : null}
            </DialogContent>
        </Dialog>
    );
}
