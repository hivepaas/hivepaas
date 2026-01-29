import React, { useEffect, useRef, useState } from "react";

import { Button } from "@components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { UsersCommands } from "~/user-management/data/commands";

import { LinkGenerate } from "../building-blocks";
import { InviteUserForm } from "../form";
import { useInviteUserDialogState } from "../hooks";
import { type InviteUserFormOutput } from "../schemas";

export function InviteUserDialog() {
    const { state, props, ...actions } = useInviteUserDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const [inviteLink, setInviteLink] = useState<string | null>(null);
    const formRef = useRef<HTMLFormElement>(null);

    const { mutate: inviteUser, isPending: isGeneratingLink } = UsersCommands.useInviteOne();

    const open = state.mode !== "closed";

    // Reset invite link when dialog closes
    useEffect(() => {
        if (state.mode === "closed") {
            setInviteLink(null);
            setHasChanges(false);
        }
    }, [state.mode]);

    function onSubmit(values: InviteUserFormOutput) {
        inviteUser(
            { user: values },
            {
                onSuccess: response => {
                    setInviteLink(response.data.inviteLink);
                },
            },
        );
    }

    function handleClose(): void {
        if (hasChanges && !isGeneratingLink) {
            const userConfirmed: boolean = window.confirm("Are you sure you want to close without saving changes?");
            if (!userConfirmed) {
                return;
            }
        }

        setHasChanges(false);
        actions.close();
    }

    return (
        <Dialog
            open={open}
            modal
            onOpenChange={handleClose}
        >
            <DialogContent className="lg:min-w-[700px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Invite a user</DialogTitle>
                    <DialogDescription>
                        Enter the required details to invite a user
                    </DialogDescription>
                </DialogHeader>
                <div className="h-px bg-border" />
                <InviteUserForm
                    ref={formRef}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                >
                    {/* Invite Link Section */}
                    <LinkGenerate inviteLink={inviteLink} />

                    {/* Footer Actions */}
                    <div className="flex items-center justify-end gap-4 pt-4 border-t">
                        <div className="flex items-center gap-4">
                            <Button
                                type="submit"
                                variant="default"
                                isLoading={isGeneratingLink}
                                disabled={inviteLink !== null}
                            >
                                Generate Invite Link
                            </Button>
                            <Button
                                type="button"
                                // variant="outline"
                                // onClick={handleSendEmail}
                            >
                                Invite via Email
                            </Button>
                        </div>
                    </div>
                </InviteUserForm>
            </DialogContent>
        </Dialog>
    );
}
