import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { ProjectAppsCommands } from "~/projects/data/commands";
import { QK } from "~/projects/data/constants";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

import { ConfirmAppDangerActionForm } from "../form";
import { useConfirmAppDangerActionDialogState } from "../hooks";
import type { ConfirmAppDangerActionFormOutput } from "../schemas";
import { AppDangerAction } from "../types";

const dialogTitle = {
    [AppDangerAction.Disable]: "Disable app",
    [AppDangerAction.ReEnable]: "Re-enable app",
    [AppDangerAction.Delete]: "Delete app",
} as const;

export function ConfirmAppDangerActionDialog() {
    const { state, props: dialogOptions, ...actions } = useConfirmAppDangerActionDialogState();
    const { navigate } = useAppNavigate();
    const queryClient = useQueryClient();
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const open = state.mode !== "closed";
    const action = state.mode === "open" ? state.action : null;
    const target = state.mode === "open" ? state.target : null;

    const { mutate: updateAppStatus, isPending: isUpdating } = ProjectAppsCommands.useUpdateStatus({
        onSuccess: (_response, request) => {
            const nextAction =
                request.payload.status === EProjectAppStatus.Active
                    ? AppDangerAction.ReEnable
                    : AppDangerAction.Disable;

            toast.success(nextAction === AppDangerAction.ReEnable ? "App re-enabled" : "App disabled");
            actions.close();
            dialogOptions?.onSuccess?.(nextAction);
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    const { mutate: deleteApp, isPending: isDeleting } = ProjectAppsCommands.useDeleteOne({
        onSuccess: (_response, request) => {
            toast.success("App deleted");
            actions.close();
            dialogOptions?.onSuccess?.(AppDangerAction.Delete);
            navigate.modules(ROUTE.projects.single.apps.$route(request.projectID), { ignorePrevPath: true });
            window.setTimeout(() => {
                queryClient.removeQueries({
                    queryKey: [
                        QK["projects.apps.$.find-one-by-id"],
                        { projectID: request.projectID, appID: request.appID },
                    ],
                    exact: true,
                });
            }, 0);
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    const isPending = isUpdating || isDeleting;

    function handleSubmit(_values: ConfirmAppDangerActionFormOutput) {
        if (!canWrite || !target || !action) {
            return;
        }

        if (action === AppDangerAction.Delete) {
            deleteApp({
                projectID: target.projectId,
                appID: target.appId,
            });
            return;
        }

        updateAppStatus({
            projectID: target.projectId,
            appID: target.appId,
            payload: {
                updateVer: target.updateVer,
                status: action === AppDangerAction.ReEnable ? EProjectAppStatus.Active : EProjectAppStatus.Disabled,
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
                        <ConfirmAppDangerActionForm
                            action={action}
                            appName={target.appName}
                            isPending={isPending}
                            onSubmit={handleSubmit}
                            readOnly={!canWrite}
                        />
                    </>
                ) : null}
            </DialogContent>
        </Dialog>
    );
}
