import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { UsersCommands } from "~/user-management/data/commands";

import { LinkGenerate } from "../building-blocks";
import { useResetUserPasswordDialogState } from "../hooks";

const fnPlaceholder = () => null;

export function ResetUserPasswordDialog() {
    const { state, props: { onClose = fnPlaceholder } = {}, ...actions } = useResetUserPasswordDialogState();
    const [resetLink, setResetLink] = useState<string | null>(null);

    const { mutate: resetPassword, isPending: isGeneratingLink } = UsersCommands.useResetPassword();

    const open = state.mode !== "closed";

    function handleResetPassword() {
        if (!state.user?.id) return;

        resetPassword(
            { id: state.user.id },
            {
                onSuccess: response => {
                    const link = response.data.resetPasswordLink;
                    setResetLink(link);
                },
            },
        );
    }

    function handleClose(): void {
        setResetLink(null);
        actions.close();
        onClose();
    }

    return (
        <Dialog
            open={open}
            modal
            onOpenChange={handleClose}
        >
            <DialogContent className="lg:min-w-[500px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Reset Password</DialogTitle>
                </DialogHeader>
                <div className="h-px bg-border" />

                <div className="flex flex-col gap-6 py-4">
                    <LinkGenerate resetLink={resetLink} />
                </div>

                <div className="flex items-center justify-end gap-4 pt-4 border-t">
                    <Button
                        type="button"
                        variant="outline"
                        onClick={handleClose}
                    >
                        Close
                    </Button>
                    <Button
                        type="button"
                        variant="default"
                        isLoading={isGeneratingLink}
                        disabled={resetLink !== null}
                        onClick={handleResetPassword}
                    >
                        Generate Reset Link
                    </Button>
                </div>
            </DialogContent>
        </Dialog>
    );
}
