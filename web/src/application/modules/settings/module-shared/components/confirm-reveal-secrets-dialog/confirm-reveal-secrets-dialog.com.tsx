import { AlertTriangle } from "lucide-react";

import { Button, Separator } from "@/components/ui";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";

export function ConfirmRevealSecretsDialog({ open, onOpenChange, onConfirm, isPending = false }: Props) {
    return (
        <Dialog
            open={open}
            onOpenChange={isPending ? undefined : onOpenChange}
        >
            <DialogFixedContent className="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle>Reveal Secrets</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>

                <DialogBody className="flex flex-col gap-4">
                    <div className="flex items-start gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive font-medium leading-normal">
                        <AlertTriangle className="size-4 shrink-0 text-destructive mt-0.5" />
                        <span>
                            Warning: Revealing secrets is not recommended for security reasons. Only administrators and
                            users with the &quot;Can Reveal Secrets&quot; capability can view secrets. This action will
                            also be recorded in the database for auditing purposes.
                        </span>
                    </div>
                </DialogBody>

                <DialogActionFooter>
                    <Button
                        type="button"
                        variant="outline"
                        onClick={() => {
                            onOpenChange(false);
                        }}
                        disabled={isPending}
                    >
                        Cancel
                    </Button>
                    <Button
                        type="button"
                        variant="destructive"
                        onClick={() => {
                            void onConfirm();
                        }}
                        isLoading={isPending}
                        disabled={isPending}
                    >
                        Reveal the secrets
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}

interface Props {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onConfirm: () => void | Promise<void>;
    isPending?: boolean;
}
