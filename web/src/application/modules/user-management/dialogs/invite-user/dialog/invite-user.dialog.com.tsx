import React, { useEffect, useRef, useState } from "react";

import { Button } from "@components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@components/ui/dialog";

import { LabelWithInfo } from "@application/shared/components";

import { LinkGenerate } from "../building-blocks";
import { InviteUserForm } from "../form";
import { useInviteUserDialogState } from "../hooks";
import { type InviteUserFormOutput } from "../schemas";

export function InviteUserDialog() {
    const { state, props, ...actions } = useInviteUserDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const [inviteLink, setInviteLink] = useState<string | null>(null);
    const [isGeneratingLink, setIsGeneratingLink] = useState(false);
    const formRef = useRef<HTMLFormElement>(null);

    const open = state.mode !== "closed";

    // Reset invite link when dialog closes
    useEffect(() => {
        if (state.mode === "closed") {
            setInviteLink(null);
            setHasChanges(false);
        }
    }, [state.mode]);

    function onSubmit(values: InviteUserFormOutput) {
        // TODO: Call Commands/Queries to invite user
        // Example:
        // inviteUser(values).then(() => {
        //     actions.close();
        //     props?.onSuccess?.();
        // });
        console.log("Invite user:", values);
        // actions.close();
        setIsGeneratingLink(true);
        setTimeout(() => {
            setInviteLink("https://example.com/invite/abc123xyz");
            setIsGeneratingLink(false);
        }, 1000);
    }

    function handleSendEmail() {
        // Trigger form submit
        formRef.current?.requestSubmit();
    }

    return (
        <Dialog
            open={open}
            modal
            onOpenChange={actions.close}
        >
            <DialogContent className="min-w-[800px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Invite a user</DialogTitle>
                    <DialogDescription>
                        Please enter the below information to invite a new user to the system
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
                    <div className="flex items-center justify-between gap-4 pt-4 border-t">
                        <Button
                            type="button"
                            variant="outline"
                            onClick={actions.close}
                        >
                            Cancel
                        </Button>
                        <div className="flex items-center gap-4">
                            <LabelWithInfo
                                className="mb-[0!important] gap-[16px!important]"
                                label={
                                    <Button
                                        type="button"
                                        // variant="outline"
                                        onClick={handleSendEmail}
                                        disabled
                                    >
                                        Send Email
                                    </Button>
                                }
                                content="Send an email to the user with the invite link"
                            />

                            <Button
                                type="submit"
                                variant="default"
                                isLoading={isGeneratingLink}
                                disabled={isGeneratingLink}
                            >
                                Generate Invite Link
                            </Button>
                        </div>
                    </div>
                </InviteUserForm>
            </DialogContent>
        </Dialog>
    );
}
