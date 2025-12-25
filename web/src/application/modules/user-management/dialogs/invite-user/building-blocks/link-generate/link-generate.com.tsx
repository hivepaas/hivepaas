import { memo } from "react";

import { Button } from "@components/ui/button";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
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

    return (
        <div className={cn(dashedBorderBox)}>
            <div className="flex items-center justify-center gap-4">
                <div className="text-sm text-foreground text-center">
                    {inviteLink ? (
                        <p className="break-all">{inviteLink}</p>
                    ) : (
                        <>
                            Press the below button to get the invitation link and
                            <br />
                            give it to the user you want to invite
                        </>
                    )}
                </div>
                <Button
                    type="button"
                    variant="link"
                    size="icon"
                    className="shrink-0"
                    disabled={!inviteLink}
                    onClick={handleCopyLink}
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
