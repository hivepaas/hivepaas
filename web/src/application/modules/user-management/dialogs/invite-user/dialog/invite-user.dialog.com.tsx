import React, { useEffect, useRef, useState } from "react";

import { Button } from "@components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { Tooltip, TooltipContent, TooltipTrigger } from "@components/ui/tooltip";
import { CircleHelp } from "lucide-react";
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

    const [sendInviteEmail, setSendInviteEmail] = useState(false);

    // Reset invite link when dialog closes
    useEffect(() => {
        if (state.mode === "closed") {
            setInviteLink(null);
            setHasChanges(false);
            setSendInviteEmail(false);
        }
    }, [state.mode]);

    function onSubmit(values: InviteUserFormOutput) {
        inviteUser(
            { user: values, sendInviteEmail },
            {
                onSuccess: response => {
                    setInviteLink(response.data.inviteLink);
                    setSendInviteEmail(false); // Reset after submission
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
                    <DialogDescription>Enter the required details to invite a user</DialogDescription>
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
                    <div className="flex items-center justify-end gap-3 pt-4 border-t">
                        <Button
                            type="submit"
                            variant="default"
                            isLoading={isGeneratingLink}
                            disabled={inviteLink !== null}
                            onClick={() => {
                                setSendInviteEmail(true);
                            }}
                        >
                            Send Email
                        </Button>
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <button
                                    type="button"
                                    className="text-muted-foreground hover:text-foreground transition-colors"
                                >
                                    <CircleHelp className="size-5" />
                                </button>
                            </TooltipTrigger>
                            <TooltipContent side="top">
                                <p className="max-w-xs">
                                    <strong>Send Email:</strong> Automatically sends an invitation email to the user
                                    <br />
                                    <strong>Generate Link:</strong> Creates an invite link you can share manually
                                </p>
                            </TooltipContent>
                        </Tooltip>
                        <Button
                            type="submit"
                            variant="default"
                            isLoading={isGeneratingLink}
                            disabled={inviteLink !== null}
                            onClick={() => {
                                setSendInviteEmail(false);
                            }}
                        >
                            Generate Invite Link
                        </Button>
                    </div>
                </InviteUserForm>
            </DialogContent>
        </Dialog>
    );
}
