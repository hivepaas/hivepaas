import React, { useState } from "react";

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { useUpdateEffect } from "react-use";
import { toast } from "sonner";

import { ProfileCommands } from "@application/shared/data/commands";

import { UpdateApiKeyStatusForm } from "../form";
import { useUpdateApiKeyStatusDialogState } from "../hooks";
import { type UpdateApiKeyStatusFormOutput } from "../schemas";

const fnPlaceholder = () => null;

export function UpdateApiKeyStatusDialog() {
    const { state, props: { onClose = fnPlaceholder } = {}, ...actions } = useUpdateApiKeyStatusDialogState();

    const [hasChanges, setHasChanges] = useState(false);

    const { mutate: updateOneApiKeyStatus, isPending: isUpdating } = ProfileCommands.useUpdateOneApiKeyStatus({});

    // Clear props when dialog closes
    useUpdateEffect(() => {
        if (state.mode === "closed") {
            actions.clear();
        }
    }, [state.mode]);

    function onSubmit(values: UpdateApiKeyStatusFormOutput) {
        if (state.mode !== "open") return;

        updateOneApiKeyStatus(
            {
                id: state.apiKey.id,
                status: values.status,
                expireAt: values.expireAt,
            },
            {
                onSuccess: () => {
                    toast.success("API key status updated successfully");
                    onClose();
                },
            },
        );
    }

    const open = state.mode !== "closed";
    const isOpen = state.mode === "open";

    // Prepare initial values from apiKey
    const initialValues =
        state.mode === "open"
            ? {
                  status: state.apiKey.status,
                  expireAt: state.apiKey.expireAt ? new Date(state.apiKey.expireAt) : undefined,
              }
            : undefined;

    function handleClose() {
        if (isUpdating) return;

        if (hasChanges) {
            const canClose = window.confirm("Are you sure you want to close modal without saving changes?");

            if (!canClose) return;
        }
        actions.close();
    }

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle>Change status</DialogTitle>
                    <DialogDescription />
                </DialogHeader>
                {isOpen && (
                    <UpdateApiKeyStatusForm
                        isPending={isUpdating}
                        onSubmit={onSubmit}
                        onHasChanges={setHasChanges}
                        initialValues={initialValues}
                    />
                )}
            </DialogContent>
        </Dialog>
    );
}
