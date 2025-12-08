import React from "react";

import { Button } from "@components/ui/button";
import { Copy } from "lucide-react";
import { toast } from "sonner";

function View({ keyId }: { keyId: string }) {
    function handleCopyKeyId() {
        void navigator.clipboard
            .writeText(keyId)
            .then(() => {
                toast.success("Key ID copied to clipboard");
            })
            .catch(() => {
                toast.error("Failed to copy Key ID");
            });
    }

    return (
        <div className="flex items-center gap-2">
            <span className="font-mono">{keyId}</span>
            <Button
                type="button"
                variant="link"
                size="icon"
                className="shrink-0"
                onClick={handleCopyKeyId}
            >
                <Copy className="h-4 w-4 text-muted-foreground" />
            </Button>
        </div>
    );
}

export const KeyIdCell = React.memo(View);
