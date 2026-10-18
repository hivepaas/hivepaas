import { useState } from "react";

import { Dialog, DialogBody, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";
import { AppManagedCommands, AppManagedQueries } from "~/projects/data";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { Separator } from "@/components/ui";

import { ManagedRestorePointForm } from "../form";
import { useManagedRestorePointDialogState } from "../hooks";
import type { ManagedRestorePointFormOutput } from "../schemas";

/** Makes a restore point in a managed database, or changes a point's note and how long it is kept. */
export function ManagedRestorePointDialog() {
    const { state, props: dialogOptions, ...actions } = useManagedRestorePointDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const open = state.mode !== "closed";
    const creating = state.mode === "create";
    const scope = {
        projectID: state.target?.projectId ?? "",
        env: state.target?.env ?? "",
        appID: state.target?.appId ?? "",
    };

    // The tab's queries: what the estimate of a kept point goes by.
    const { data: settings, isLoading: isLoadingSettings } = AppManagedQueries.useFindSettings(scope, {
        enabled: open,
    });
    const { data: status, isLoading: isLoadingStatus } = AppManagedQueries.useFindStatus(scope, { enabled: open });

    const onDone = () => {
        setHasChanges(false);
        actions.close();
        dialogOptions?.onSuccess?.();
    };
    const { mutate: createPoint, isPending: isCreating } = AppManagedCommands.useCreateRestorePoint({
        onSuccess: response => {
            toast.success(`Restore point ${response.data.name} made`);
            onDone();
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });
    const { mutate: updatePoint, isPending: isUpdating } = AppManagedCommands.useUpdateRestorePoint({
        onSuccess: () => {
            toast.success("Restore point updated");
            onDone();
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });
    const isPending = isCreating || isUpdating;

    function onSubmit(values: ManagedRestorePointFormOutput) {
        if (!canWrite) {
            return;
        }

        const keepUntil = values.keepUntil ? values.keepUntil.toISOString() : null;
        if (state.mode === "create") {
            createPoint({
                ...scope,
                payload: {
                    name: values.name,
                    ...(values.note !== "" ? { note: values.note } : {}),
                    ...(keepUntil ? { keepUntil } : {}),
                },
            });
        } else if (state.mode === "edit") {
            updatePoint({ ...scope, name: state.point.name, payload: { keepUntil, note: values.note } });
        }
    }

    function handleClose(): void {
        if (isPending) {
            return;
        }

        if (canWrite && hasChanges) {
            const userConfirmed = window.confirm("Are you sure you want to close without saving changes?");
            if (!userConfirmed) {
                return;
            }
        }

        setHasChanges(false);
        actions.close();
        dialogOptions?.onClose?.();
    }

    const initialValues =
        state.mode === "edit"
            ? { name: state.point.name, note: state.point.note, keepUntil: state.point.keepUntil ?? null }
            : undefined;

    return (
        <Dialog
            open={open}
            onOpenChange={isOpen => {
                if (!isOpen) {
                    handleClose();
                }
            }}
        >
            <DialogFixedContent className="sm:max-w-[640px]">
                <DialogHeader>
                    <DialogTitle>{creating ? "Make a restore point" : "Edit restore point"}</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                {isLoadingSettings || isLoadingStatus ? (
                    <DialogBody>
                        <AppLoader />
                    </DialogBody>
                ) : (
                    <ManagedRestorePointForm
                        // A fresh form for each point.
                        key={state.mode === "edit" ? state.point.name : state.mode}
                        creating={creating}
                        initialValues={initialValues}
                        retainFull={settings?.data.retainFull ?? 0}
                        lastBackupBytes={status?.data.lastBackup?.bytes}
                        isPending={isPending}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        readOnly={!canWrite}
                    />
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
