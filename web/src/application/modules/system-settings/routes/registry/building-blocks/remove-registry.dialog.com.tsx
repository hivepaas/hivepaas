import { useEffect, useState } from "react";

import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { AlertTriangle } from "lucide-react";
import { ConfirmDangerTargetBadge } from "~/projects/module-shared/components";

import { Button, Checkbox, Input, Separator } from "@/components/ui";

interface Props {
    open: boolean;
    /** What the operator types to confirm, and what identifies the registry here. */
    domain: string;
    /** Images in a bucket are not ours to delete, so the choice is not offered. */
    onS3: boolean;
    isPending: boolean;
    onOpenChange: (open: boolean) => void;
    onConfirm: (removeStorage: boolean) => void;
}

/**
 * Switching the registry off takes its app down, and there is no other screen to
 * do it from: the project the registry lives in is left out of every listing. So
 * the confirmation that would have lived on the app's own page lives here.
 */
export function RemoveRegistryDialog({ open, domain, onS3, isPending, onOpenChange, onConfirm }: Props) {
    const [typed, setTyped] = useState("");
    const [removeStorage, setRemoveStorage] = useState(false);

    useEffect(() => {
        if (open) {
            setTyped("");
            setRemoveStorage(false);
        }
    }, [open]);

    const confirmed = typed.trim() === domain;

    return (
        <Dialog
            open={open}
            onOpenChange={nextOpen => {
                if (isPending) {
                    return;
                }
                onOpenChange(nextOpen);
            }}
        >
            <DialogFixedContent className="sm:max-w-[560px]">
                <DialogHeader>
                    <DialogTitle>Switch the registry off</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <DialogBody className="flex flex-col gap-4">
                    <p className="text-sm leading-6">
                        This removes the app that serves the registry and the service it runs. Apps that pull their
                        images from it keep running, but their next deployment has nowhere to push to.
                    </p>

                    <div className="flex items-center gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm font-medium text-destructive">
                        <AlertTriangle className="size-4 shrink-0" />
                        <span>Warning: this cannot be undone.</span>
                    </div>

                    {onS3 ? (
                        <p className="text-sm leading-6 text-muted-foreground">
                            The images stay in the bucket. Nothing here reaches into it, so deleting them is done where
                            the bucket is.
                        </p>
                    ) : (
                        <label
                            className="flex items-start gap-2.5 text-sm leading-6"
                            htmlFor="registry-remove-storage"
                        >
                            <Checkbox
                                id="registry-remove-storage"
                                className="mt-1"
                                checked={removeStorage}
                                disabled={isPending}
                                onCheckedChange={checked => {
                                    setRemoveStorage(checked === true);
                                }}
                            />
                            <span>
                                <span className="block">Also delete the images</span>
                            </span>
                        </label>
                    )}

                    <div className="flex flex-col gap-2">
                        <p className="text-sm font-medium leading-6">
                            Type <ConfirmDangerTargetBadge text={domain} /> to confirm.
                        </p>
                        <Input
                            value={typed}
                            disabled={isPending}
                            placeholder={domain}
                            onChange={event => {
                                setTyped(event.target.value);
                            }}
                        />
                    </div>
                </DialogBody>
                <DialogActionFooter>
                    <Button
                        variant="outline"
                        disabled={isPending}
                        onClick={() => {
                            onOpenChange(false);
                        }}
                    >
                        Cancel
                    </Button>
                    <Button
                        variant="destructive"
                        disabled={!confirmed || isPending}
                        isLoading={isPending}
                        onClick={() => {
                            onConfirm(removeStorage);
                        }}
                    >
                        Switch off and remove
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
