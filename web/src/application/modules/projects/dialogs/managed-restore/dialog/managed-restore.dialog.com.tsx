import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { type AppManagedRestorePayload } from "~/projects/api/services";
import { AppManagedCommands, AppManagedQueries, ProjectClusterVolumesQueries } from "~/projects/data";
import { MANAGED_RESTORE_TARGET } from "~/projects/domain";
import { isVolumeInMemory, volumePlaceOf } from "~/projects/module-shared/utils";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { Separator } from "@/components/ui";

import { ManagedRestoreForm, type ManagedRestoreVolumeOption } from "../form";
import { useManagedRestoreDialogState } from "../hooks";
import type { ManagedRestoreFormOutput } from "../schemas";

const LIST_ALL_PAGE = { page: 1, size: 1000 };

/**
 * Restores a managed database into a new app: to the latest it archived, a
 * time, a restore point or a base backup. The new app recovers in a task, which
 * the dialog goes to.
 */
export function ManagedRestoreDialog() {
    const navigate = useNavigate();
    const { state, props: dialogOptions, ...actions } = useManagedRestoreDialogState();
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const open = state.mode === "open";
    const scope = {
        projectID: state.source?.projectId ?? "",
        env: state.source?.env ?? "",
        appID: state.source?.appId ?? "",
    };

    const { data: status, isLoading: isLoadingStatus } = AppManagedQueries.useFindStatus(scope, { enabled: open });
    const { data: volumesData, isLoading: isLoadingVolumes } = ProjectClusterVolumesQueries.useFindManyPaginated(
        { projectID: scope.projectID, env: scope.env, pagination: LIST_ALL_PAGE },
        { enabled: open },
    );
    const volumes: ManagedRestoreVolumeOption[] = (volumesData?.data ?? []).map(volume => ({
        id: volume.id,
        name: volume.name,
        place: volumePlaceOf(volume),
        shared: !volume.nodeId && !volume.nodeLabel,
        inMemory: isVolumeInMemory(volume),
    }));

    const { mutate: restore, isPending } = AppManagedCommands.useRestore({
        onSuccess: response => {
            toast.success("Restore started: the new app recovers in a task");
            actions.close();
            dialogOptions?.onSuccess?.();
            void navigate(
                ROUTE.projects.single.apps.single.tasks.details.$route(
                    scope.projectID,
                    scope.env,
                    response.data.app.id,
                    response.data.task.id,
                ),
            );
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    function onSubmit(values: ManagedRestoreFormOutput) {
        if (!canWrite || !open) {
            return;
        }

        const target: AppManagedRestorePayload["target"] = { kind: values.kind };
        if (values.kind === MANAGED_RESTORE_TARGET.time && values.time) {
            target.time = values.time.toISOString();
        } else if (values.kind === MANAGED_RESTORE_TARGET.point) {
            target.point = values.point;
        } else if (values.kind === MANAGED_RESTORE_TARGET.backup) {
            target.backup = values.backup;
        }
        restore({
            ...scope,
            payload: {
                target,
                dataVolume: { id: values.dataVolume },
                ...(values.name !== "" ? { name: values.name } : {}),
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
            onOpenChange={isOpen => {
                if (!isOpen) {
                    handleClose();
                }
            }}
        >
            <DialogFixedContent className="sm:max-w-[720px]">
                <DialogHeader>
                    <DialogTitle>Restore {state.source?.appName ?? "the database"}</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                {isLoadingStatus || isLoadingVolumes || !status || !state.source ? (
                    <DialogBody>
                        <AppLoader />
                    </DialogBody>
                ) : (
                    <ManagedRestoreForm
                        key={`${state.source.appId}:${state.preset?.kind ?? ""}:${state.preset?.point ?? ""}`}
                        sourceName={state.source.appName}
                        status={status.data}
                        volumes={volumes}
                        preset={state.preset}
                        isPending={isPending}
                        onSubmit={onSubmit}
                        readOnly={!canWrite}
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
