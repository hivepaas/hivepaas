import React, { useRef } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import type { AppStorageMount } from "~/projects/domain";

import { StorageMountForm } from "../form";
import { formValuesToMount } from "../form/storage-mount.form-mappers";
import { useStorageMountDialogState } from "../hooks";
import type { StorageMountFormOutput } from "../schemas";
import type { StorageMountFormRef } from "../types";

const fnPlaceholder = () => null;

function mountWithoutId(mount: AppStorageMount & { _id: string }): AppStorageMount {
    const { _id: _unused, ...rest } = mount;
    return rest;
}

export function StorageMountDialog() {
    const {
        state,
        props: { onSubmit = fnPlaceholder, onClose = fnPlaceholder } = {},
        ...actions
    } = useStorageMountDialogState();

    const formRef = useRef<StorageMountFormRef>(null);

    const open = state.mode !== "closed";
    const isEdit = state.mode === "edit";
    const projectRules = state.mode !== "closed" ? state.projectRules : undefined;
    const projectKey = state.mode !== "closed" ? state.projectKey : undefined;
    const appLocalKey = state.mode !== "closed" ? state.appLocalKey : undefined;
    const defaultValues = state.mode === "edit" ? mountWithoutId(state.mount) : undefined;

    function handleSubmit(values: StorageMountFormOutput) {
        onSubmit(formValuesToMount(values));
        actions.close();
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
            <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>{isEdit ? "Edit Storage Mount" : "Add Storage Mount"}</DialogTitle>
                </DialogHeader>

                {open && (
                    <StorageMountForm
                        ref={formRef}
                        isPending={false}
                        onSubmit={handleSubmit}
                        defaultValues={defaultValues}
                        projectRules={projectRules}
                        projectKey={projectKey}
                        appLocalKey={appLocalKey}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
