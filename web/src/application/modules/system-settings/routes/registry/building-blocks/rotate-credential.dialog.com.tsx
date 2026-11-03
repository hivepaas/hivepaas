import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";

import { Button, Separator } from "@/components/ui";

interface Props {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onConfirm: () => void;
    isPending: boolean;
}

/**
 * Rotating is safe but not free: a swarm service presents the credential it was
 * deployed with, so apps that are not redeployed before the grace period ends
 * lose the registry on their next reschedule. That is what this says.
 */
export function RotateRegistryCredentialDialog({ open, onOpenChange, onConfirm, isPending }: Props) {
    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogFixedContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>Rotate the registry password</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <DialogBody>
                    <div className="flex flex-col gap-3 text-sm">
                        <p>
                            A new password is issued and every build starts using it at once. The previous one keeps
                            working for two weeks, because a service carries the credential it was deployed with.
                        </p>
                        <p>
                            Redeploy your apps before that window ends. An app that is not redeployed keeps running, but
                            the next time one of its tasks is rescheduled onto a node without the image, the pull fails.
                        </p>
                    </div>
                </DialogBody>
                <DialogActionFooter>
                    <Button
                        variant="outline"
                        onClick={() => {
                            onOpenChange(false);
                        }}
                        disabled={isPending}
                    >
                        Cancel
                    </Button>
                    <Button
                        onClick={onConfirm}
                        disabled={isPending}
                        isLoading={isPending}
                    >
                        Rotate password
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
