import React, { useCallback, useState } from "react";

import { cn } from "@/lib/utils";
import { Check, Copy } from "lucide-react";
import { toast } from "sonner";

interface Props {
    text: string;
    className?: string;
}

export function ConfirmDangerTargetBadge({ text, className }: Props) {
    const [copied, setCopied] = useState(false);

    const handleCopy = useCallback(
        (e: React.MouseEvent<HTMLButtonElement>) => {
            e.preventDefault();
            e.stopPropagation();

            if (!text) {
                return;
            }

            void navigator.clipboard
                .writeText(text)
                .then(() => {
                    setCopied(true);
                    toast.success("Copied to clipboard");
                    setTimeout(() => {
                        setCopied(false);
                    }, 1500);
                })
                .catch(() => {
                    toast.error("Failed to copy to clipboard");
                });
        },
        [text],
    );

    return (
        <span
            className={cn(
                "inline-flex max-w-full items-center gap-1.5 rounded-[4px] border border-input bg-muted/50 px-2 py-0 align-middle font-semibold text-primary break-all",
                className,
            )}
        >
            <span>{text}</span>
            <button
                type="button"
                onClick={handleCopy}
                title="Copy to clipboard"
                aria-label="Copy to clipboard"
                className="inline-flex shrink-0 items-center justify-center rounded p-0.5 text-muted-foreground transition-colors hover:bg-background/80 hover:text-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring cursor-pointer"
            >
                {copied ? <Check className="h-3.5 w-3.5 text-emerald-500" /> : <Copy className="h-3.5 w-3.5" />}
            </button>
        </span>
    );
}
