import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { AlertTriangle } from "lucide-react";
import type { AppStorageFinding } from "~/projects/api/services";

import { Button, Separator } from "@/components/ui";

interface Props {
    open: boolean;
    findings: AppStorageFinding[];
    isPending: boolean;
    onOpenChange: (open: boolean) => void;
    /** resetStorage: true clears those directories before the mounts are made. */
    onConfirm: (resetStorage: boolean) => void;
}

/**
 * What is already in the directory a mount being added would reach.
 *
 * It is there because an app of the same key ran here before: the directory is
 * named after the app, not after the app's id, so a rebuilt app inherits what
 * the old one left. For a database that is a cluster whose password nobody has
 * any more - the new app is given a freshly generated one.
 */
export function StorageInUseDialog({ open, findings, isPending, onOpenChange, onConfirm }: Props) {
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
                    <DialogTitle>This storage already has data</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <DialogBody className="flex flex-col gap-4">
                    <p className="text-sm leading-6">
                        {findings.length === 1 ? "The directory" : "The directories"} this app would be given
                        {findings.length === 1 ? " is" : " are"} not empty. The app will start with what is in
                        {findings.length === 1 ? " it" : " them"}.
                    </p>

                    <ul className="flex flex-col gap-1.5 rounded-md border border-border/80 bg-muted/30 px-3.5 py-2.5">
                        {findings.map(finding => (
                            <li
                                key={`${finding.target}:${finding.path}`}
                                className="flex flex-wrap items-baseline gap-x-2 text-sm"
                            >
                                <span className="font-mono text-xs">{finding.target}</span>
                                <span className="font-mono text-xs text-muted-foreground">
                                    {finding.volume.name}/{finding.path}
                                </span>
                            </li>
                        ))}
                    </ul>

                    <div className="flex items-start gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive">
                        <AlertTriangle className="mt-0.5 size-4 shrink-0" />
                        <span>
                            A database started on data it did not create keeps the password that data was made with, and
                            will not open with the one this app has.
                        </span>
                    </div>

                    <p className="text-xs text-muted-foreground">
                        Deleting removes only {findings.length === 1 ? "that directory" : "those directories"}. The
                        volume, and anything else on it, stays.
                    </p>
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
                        variant="outline"
                        disabled={isPending}
                        onClick={() => {
                            onConfirm(false);
                        }}
                    >
                        Save anyway
                    </Button>
                    <Button
                        variant="destructive"
                        disabled={isPending}
                        isLoading={isPending}
                        onClick={() => {
                            onConfirm(true);
                        }}
                    >
                        Delete that data and save
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
