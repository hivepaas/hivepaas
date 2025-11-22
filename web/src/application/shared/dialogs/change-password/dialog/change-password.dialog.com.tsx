import React from "react";

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { toast } from "sonner";

import { ProfileCommands } from "@application/shared/data/commands";

import { ChangePasswordForm } from "../form";
import { useChangePasswordDialogState } from "../hooks";
import { type AccountPasswordFormSchemaOutput } from "../schemas";

const fnPlaceholder = () => null;

export function ChangePasswordDialog() {
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

    const open = state.mode !== "closed";

    return (
        <Dialog
            open={open}
            onOpenChange={actions.close}
        >
            <DialogContent className="min-w-[400px] w-fit">
                <DialogHeader>
                    <DialogTitle>Change Password</DialogTitle>
                    <DialogDescription />
                </DialogHeader>
                <ChangePasswordForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                />
            </DialogContent>
        </Dialog>
    );
}
