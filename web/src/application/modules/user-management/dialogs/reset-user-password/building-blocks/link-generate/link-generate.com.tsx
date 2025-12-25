import { memo } from "react";

import { Button } from "@components/ui/button";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { Copy } from "lucide-react";
import { toast } from "sonner";

function View({ resetLink }: Props) {
    function handleCopyLink() {
        if (!resetLink) {
            return;
        }

        void navigator.clipboard
            .writeText(resetLink)
            .then(() => {
                toast.success("Reset link copied to clipboard");
            })
            .catch(() => {
                toast.error("Failed to copy link");
            });
    }

    return (
        <div className={cn(dashedBorderBox)}>
            <div className="flex items-center justify-center gap-4">
                <div className="text-sm text-foreground text-center">
                    {resetLink ? (
                        <p className="break-all">{resetLink}</p>
                    ) : (
                        <>Press the below button to get the password reset link and give it to the user</>
                    )}
                </div>
                <Button
                    type="button"
                    variant="link"
                    size="icon"
                    className="shrink-0"
                    disabled={!resetLink}
                    onClick={handleCopyLink}
                >
                    <Copy className="h-4 w-4 text-muted-foreground" />
                </Button>
            </div>
        </div>
    );
}

interface Props {
    resetLink: string | null;
}

export const LinkGenerate = memo(View);
