import React, { useEffect, useRef, useState } from "react";

import { Button } from "@components/ui/button";
import {
    Dialog,
    DialogActionFooter,
    DialogDescription,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { UsersCommands } from "~/user-management/data/commands";

import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { LinkGenerate } from "../building-blocks";
import { InviteUserForm } from "../form";
import { useInviteUserDialogState } from "../hooks";
import { type InviteUserFormOutput } from "../schemas";

export function InviteUserDialog() {
    const { state, props, ...actions } = useInviteUserDialogState();
    const [hasChanges, setHasChanges] = useState(false);
    const [inviteLink, setInviteLink] = useState<string | null>(null);
    const formRef = useRef<HTMLFormElement>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.User });

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
        if (!canWrite) {
            return;
        }

        inviteUser(
            { user: values, sendInviteEmail },
            {
                onSuccess: response => {
                    if (sendInviteEmail) {
                        setInviteLink(null);
                    } else {
                        setInviteLink(response.data.inviteLink);
                    }
                    setSendInviteEmail(false); // Reset after submission
                },
            },
        );
    }

    function handleClose(): void {
        if (canWrite && hasChanges && !isGeneratingLink) {
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
            <DialogFixedContent className="lg:min-w-[700px]">
                <DialogHeader>
                    <DialogTitle>Invite a user</DialogTitle>
                    <DialogDescription>Enter the required details to invite a user</DialogDescription>
                </DialogHeader>
                <InviteUserForm
                    ref={formRef}
                    readOnly={!canWrite}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                    footer={
                        <DialogActionFooter className="flex items-center justify-end gap-3">
                            <PermissionTooltipAction
                                id={MODULE_IDS.User}
                                action="write"
                            >
                                {({ isDenied }) => (
                                    <Button
                                        type="submit"
                                        variant="default"
                                        isLoading={isGeneratingLink}
                                        disabled={isDenied}
                                        onClick={() => {
                                            setSendInviteEmail(true);
                                        }}
                                    >
                                        Send Email
                                    </Button>
                                )}
                            </PermissionTooltipAction>
                            <PermissionTooltipAction
                                id={MODULE_IDS.User}
                                action="write"
                            >
                                {({ isDenied }) => (
                                    <Button
                                        type="submit"
                                        variant="default"
                                        isLoading={isGeneratingLink}
                                        disabled={isDenied || inviteLink !== null}
                                        onClick={() => {
                                            setSendInviteEmail(false);
                                        }}
                                    >
                                        Generate Invite Link
                                    </Button>
                                )}
                            </PermissionTooltipAction>
                        </DialogActionFooter>
                    }
                >
                    {/* Invite Link Section */}
                    <LinkGenerate inviteLink={inviteLink} />
                </InviteUserForm>
            </DialogFixedContent>
        </Dialog>
    );
}
