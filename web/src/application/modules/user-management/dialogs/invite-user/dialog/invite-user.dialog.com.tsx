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
import { Send } from "lucide-react";
import { toast } from "sonner";
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

    // Which button was pressed, kept in a ref rather than state because it is
    // written on click and read on the submit that the same click triggers, and
    // never rendered. Reading a stale value here would close the dialog on the
    // "Generate Invite Link" path and take the freshly generated link with it.
    const sendInviteEmailRef = useRef(false);

    // Reset invite link when dialog closes
    useEffect(() => {
        if (state.mode === "closed") {
            setInviteLink(null);
            setHasChanges(false);
            sendInviteEmailRef.current = false;
        }
    }, [state.mode]);

    function onSubmit(values: InviteUserFormOutput) {
        if (!canWrite) {
            return;
        }

        const sendInviteEmail = sendInviteEmailRef.current;

        inviteUser(
            { user: values, sendInviteEmail },
            {
                onSuccess: response => {
                    // The generated link is shown nowhere else, so that path has to
                    // keep the dialog open. The email path has nothing left to show,
                    // and leaving it open reads as if the invite had not been sent.
                    if (!sendInviteEmail) {
                        setInviteLink(response.data.inviteLink);
                        return;
                    }

                    toast.success(`An invitation email has been sent to ${values.email}.`);
                    actions.close();
                },
                // Failures need no handling here: useUsersApi.inviteOne reports them
                // through notifyError, which raises the toast. The dialog stays open
                // with what was typed still in it, ready to be corrected.
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
            <DialogFixedContent className="sm:max-w-[850px]">
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
                                            sendInviteEmailRef.current = true;
                                        }}
                                    >
                                        {/* Hidden while loading: Button prepends its own spinner, and the
                                            two icons would otherwise sit side by side. */}
                                        {!isGeneratingLink && <Send className="size-4" />}
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
                                            sendInviteEmailRef.current = false;
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
