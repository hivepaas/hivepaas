import React, { memo } from "react";

import { Button } from "@components/ui/button";
import { Input } from "@components/ui/input";
import { Copy } from "lucide-react";
import { toast } from "sonner";

function View({ inviteLink }: Props) {
    function handleCopyLink() {
        if (!inviteLink) {
            return;
        }

        void navigator.clipboard
            .writeText(inviteLink)
            .then(() => {
                toast.success("Invite link copied to clipboard");
            })
            .catch(() => {
                toast.error("Failed to copy link");
            });
    }
    if (inviteLink) {
        return (
            <div className="flex flex-col gap-2">
                <label
                    htmlFor="invite-link-input"
                    className="text-sm font-medium"
                >
                    Invitation Link
                </label>
                <div className="relative">
                    <Input
                        id="invite-link-input"
                        type="text"
                        value={inviteLink}
                        readOnly
                        className="pr-10 border border-green-500 bg-green-50/50 "
                    />
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="absolute right-0 top-0 h-full px-3 hover:bg-transparent"
                        onClick={() => {
                            handleCopyLink();
                        }}
                    >
                        <Copy className="h-4 w-4" />
                        <span className="sr-only">Copy invite link</span>
                    </Button>
                </div>
            </div>
        );
    }

    return (
        <div className="border border-dashed border-purple-500 rounded-lg p-6 bg-purple-50/50 dark:bg-purple-950/20">
            <div className="flex items-center justify-center gap-4">
                <p className="text-sm text-foreground text-center">
                    Press the below button to get the invitation link and
                    <br />
                    give it to the user you want to invite
                </p>
                <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="shrink-0"
                    disabled
                >
                    <Copy className="h-4 w-4 text-muted-foreground" />
                </Button>
            </div>
        </div>
    );
}

interface Props {
    inviteLink: string | null;
}

export const LinkGenerate = memo(View);
