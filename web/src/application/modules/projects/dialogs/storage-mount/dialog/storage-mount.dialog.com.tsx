import React from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";

import { StorageMountForm } from "../form";
import { useStorageMountDialogState } from "../hooks";
import type { StorageMountFormInput, StorageMountFormOutput } from "../schemas";

const fnPlaceholder = () => null;
export function StorageMountDialog() {
    const {
        state,
        props: { onSubmit = fnPlaceholder, onClose = fnPlaceholder } = {},
        ...actions
    } = useStorageMountDialogState();

    const open = state.mode !== "closed";
    const isEdit = state.mode === "edit";
    const projectRules = state.mode !== "closed" ? state.projectRules : undefined;
    const initialValues = state.mode === "edit" ? (state.mount as StorageMountFormInput) : undefined;

    function handleSubmit(mount: StorageMountFormOutput) {
        onSubmit(mount);
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
                        isPending={false}
                        onSubmit={handleSubmit}
                        initialValues={initialValues}
                        projectRules={projectRules}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
