import React, { useRef, useState } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import type { AppStorageMount } from "~/projects/domain";

import { StorageMountForm } from "../form";
import { formValuesToMount } from "../form/storage-mount.form-mappers";
import { useStorageMountDialogState } from "../hooks";
import type { StorageMountFormOutput } from "../schemas";
import type { StorageMountFormRef } from "../types";

const fnPlaceholder = () => null;
const errorPlaceholder = (_error: Error) => null;

function mountWithoutId(mount: AppStorageMount & { _id: string }): AppStorageMount {
    const { _id: _unused, ...rest } = mount;
    return rest;
}

export function StorageMountDialog() {
    const {
        state,
        props: { onSubmit = fnPlaceholder, onClose = fnPlaceholder, onError = errorPlaceholder } = {},
        ...actions
    } = useStorageMountDialogState();

    const formRef = useRef<StorageMountFormRef>(null);
    const [isSubmitting, setIsSubmitting] = useState(false);

    const open = state.mode !== "closed";
    const isEdit = state.mode === "edit";
    const projectKey = state.mode !== "closed" ? state.projectKey : undefined;
    const appLocalKey = state.mode !== "closed" ? state.appLocalKey : undefined;
    const defaultValues = state.mode === "edit" ? mountWithoutId(state.mount) : undefined;

    async function handleSubmit(values: StorageMountFormOutput) {
        try {
            setIsSubmitting(true);
            await onSubmit(formValuesToMount(values));
            actions.close();
        } catch (error) {
            const handledError = error instanceof Error ? error : new Error("Failed to save storage mount");
            onError(handledError);
        } finally {
            setIsSubmitting(false);
        }
    }

    function handleClose() {
        actions.close();
        onClose();
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
            <DialogContent className="min-w-[390px] w-[680px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>{isEdit ? "Edit storage" : "Add a new storage to the app"}</DialogTitle>
                </DialogHeader>

                {open && (
                    <>
                        <div className={cn(dashedBorderBox, "text-[12px] text-center")}>
                            <span className="font-bold text-orange-500">Important:</span> If your cluster consists of
                            more than 1 node, you need to ensure that the directories or volumes are accessible from all
                            nodes. Otherwise, your apps may not function properly.
                        </div>
                        <StorageMountForm
                            ref={formRef}
                            isPending={isSubmitting}
                            onSubmit={values => void handleSubmit(values)}
                            defaultValues={defaultValues}
                            projectKey={projectKey}
                            appLocalKey={appLocalKey}
                        />
                    </>
                )}
            </DialogContent>
        </Dialog>
    );
}
