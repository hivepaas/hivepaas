import React, { useState } from "react";

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";

import { ProfileCommands } from "@application/shared/data/commands";

import { ChangePasswordForm } from "../form";
import { useChangePasswordDialogState } from "../hooks";
import { type AccountPasswordFormSchemaOutput } from "../schemas";

const fnPlaceholder = () => null;

export function ChangePasswordDialog() {
    const [hasChanges, setHasChanges] = useState(false);
    const { state, props: { onClose = fnPlaceholder } = {}, ...actions } = useChangePasswordDialogState();

    const { mutate: updatePassword, isPending } = ProfileCommands.useUpdatePassword();

    function onSubmit(values: AccountPasswordFormSchemaOutput) {
        updatePassword(
            {
                currentPassword: values.currentPassword,
                newPassword: values.newPassword,
            },
            {
                onSuccess: () => {
                    toast.success("Password updated successfully");
                    onClose();
                },
            },
        );
    }

    function handleClose() {
        if (hasChanges) {
            const canClose = window.confirm("Are you sure you want to close modal without saving changes?");

            if (!canClose) return;
        }
        actions.close();
    }
    const open = state.mode !== "closed";

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogContent className="min-w-[400px] w-fit">
                <DialogHeader>
                    <DialogTitle>Change Password</DialogTitle>
                    <DialogDescription />
                </DialogHeader>
                <ChangePasswordForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                />
            </DialogContent>
        </Dialog>
    );
}
