import { ClipboardCopy } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui";

/**
 * The "Copy ID" entry of a table row's actions menu: it copies the row's ID, the
 * one the API and the MCP tools take, and tells the menu it is done.
 */
export function CopyIdMenuButton({ id, onCopied }: Props) {
    function handleClick() {
        navigator.clipboard.writeText(id).then(
            () => {
                toast.success("ID copied to clipboard");
            },
            () => {
                toast.error("Failed to copy ID");
            },
        );
        onCopied?.();
    }

    return (
        <Button
            className="justify-start py-1.5 w-full"
            variant="ghost"
            onClick={handleClick}
        >
            <ClipboardCopy className="mr-2 size-4" />
            Copy ID
        </Button>
    );
}

interface Props {
    id: string;
    /** Called once the copy is started, for the menu to close. */
    onCopied?: () => void;
}
