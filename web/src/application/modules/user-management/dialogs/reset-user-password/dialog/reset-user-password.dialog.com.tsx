import React, { useState } from "react";

import { Button } from "@components/ui/button";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
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
            <DialogFixedContent className="lg:min-w-[500px] lg:max-w-[700px]">
                <DialogHeader>
                    <DialogTitle>Reset Password</DialogTitle>
                </DialogHeader>

                <DialogBody className="flex flex-col gap-6">
                    <LinkGenerate resetLink={resetLink} />
                </DialogBody>

                <DialogActionFooter className="flex items-center justify-end gap-4">
                    <Button
                        type="button"
                        className="min-w-[100px]"
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
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
