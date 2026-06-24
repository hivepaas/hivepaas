import { useMemo, useState } from "react";

import { Dialog, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { ProjectAppsCommands } from "~/projects/data/commands";
import { ProjectAppsQueries, ProjectsQueries } from "~/projects/data/queries";
import {
    PROJECT_ENV_FILTER_ALL,
    useProjectEnvFilterStore,
    useSelectedProjectEnv,
} from "~/projects/module-shared/hooks";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { CopyProjectAppForm, mapCopyProjectAppFormToPayload, mapPreparedCopyToFormInput } from "../form";
import { useCopyProjectAppDialogState } from "../hooks";
import type { CopyProjectAppFormOutput } from "../schemas";

export function CopyProjectAppDialog() {
    const { state, props: dialogOptions, ...actions } = useCopyProjectAppDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const open = state.mode !== "closed";
    const projectId = state.projectId ?? "";
    const appId = state.appId ?? "";
    const selectedEnv = useSelectedProjectEnv(projectId);
    const setSelectedEnv = useProjectEnvFilterStore(store => store.setSelectedEnv);

    const { data: projectData } = ProjectsQueries.useFindOneById(
        { projectID: projectId },
        { enabled: open && Boolean(projectId) },
    );
    const prepareQuery = ProjectAppsQueries.usePrepareCopy(
        { projectID: projectId, appID: appId },
        { enabled: open && Boolean(projectId && appId) },
    );
    const preparedCopy = prepareQuery.data?.data;
    const defaultValues = useMemo(() => {
        return preparedCopy ? mapPreparedCopyToFormInput(preparedCopy) : null;
    }, [preparedCopy]);
    const projectEnvs = projectData?.data.envs ?? [];

    const { mutate: copyProjectApp, isPending } = ProjectAppsCommands.useCopy({
        onSuccess: (_response, request) => {
            toast.success("Project app copied successfully");

            if (selectedEnv !== PROJECT_ENV_FILTER_ALL) {
                setSelectedEnv(request.projectID, request.targetEnv || PROJECT_ENV_FILTER_ALL);
            }

            setHasChanges(false);
            actions.close();
            dialogOptions?.onSuccess?.();
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    function handleClose(): void {
        if (canWrite && hasChanges) {
            const userConfirmed = window.confirm("Are you sure you want to close without copying this app?");

            if (!userConfirmed) {
                return;
            }
        }

        setHasChanges(false);
        actions.close();
        dialogOptions?.onClose?.();
    }

    function handleOpenChange(nextOpen: boolean): void {
        if (nextOpen) {
            return;
        }

        handleClose();
    }

    function onSubmit(values: CopyProjectAppFormOutput) {
        if (!projectId || !appId || !canWrite) {
            return;
        }

        copyProjectApp({
            projectID: projectId,
            appID: appId,
            ...mapCopyProjectAppFormToPayload(values),
        });
    }

    if (!projectId || !appId) {
        return null;
    }

    return (
        <Dialog
            open={open}
            onOpenChange={handleOpenChange}
        >
            <DialogFixedContent className="w-[960px] max-w-[calc(100vw-2rem)]">
                <DialogHeader>
                    <DialogTitle>Copy app: {preparedCopy?.sourceName ?? ""}</DialogTitle>
                </DialogHeader>

                {!defaultValues ? (
                    <div className="flex min-h-[280px] items-center justify-center">
                        <AppLoader />
                    </div>
                ) : (
                    <CopyProjectAppForm
                        projectId={projectId}
                        envs={projectEnvs}
                        defaultValues={defaultValues}
                        isPending={isPending}
                        readOnly={!canWrite}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
