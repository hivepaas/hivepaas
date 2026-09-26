import { cn } from "@lib/utils";
import { Copy } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui";

interface Props {
    value: string;
    /** What the toast calls it once copied. */
    what: string;
    className?: string;
}

/** A value to paste somewhere else, shown as it will be pasted. */
export function CopyField({ value, what, className }: Props) {
    function handleCopy() {
        void navigator.clipboard
            .writeText(value)
            .then(() => {
                toast.success(`${what} copied`);
            })
            .catch(() => {
                toast.error(`Failed to copy the ${what.toLowerCase()}`);
            });
    }

    return (
        <div className={cn("flex items-start gap-2", className)}>
            <pre className="min-w-0 flex-1 overflow-x-auto whitespace-pre-wrap break-all rounded-md border bg-muted/40 px-3 py-2 font-mono text-xs leading-relaxed">
                {value}
            </pre>
            <Button
                type="button"
                variant="outline"
                size="icon"
                className="size-8 shrink-0"
                onClick={handleCopy}
                aria-label={`Copy the ${what.toLowerCase()}`}
            >
                <Copy className="size-4" />
            </Button>
        </div>
    );
}
