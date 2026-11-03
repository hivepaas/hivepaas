import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { AlertTriangle } from "lucide-react";

import { Button, Separator } from "@/components/ui";

import type { PreflightIssue, PreflightStorageFinding } from "../../../api";

interface Props {
    open: boolean;
    findings: PreflightStorageFinding[];
    /** What the creation would refuse. With any of these, creating is pointless. */
    issues: PreflightIssue[];
    isPending: boolean;
    onOpenChange: (open: boolean) => void;
    /** resetStorage: true deletes those directories before the apps are created. */
    onConfirm: (resetStorage: boolean) => void;
}

/**
 * What a previous install of these apps left behind.
 *
 * One decision about all of them, however many there are: a template creates up
 * to eight apps at once, and nobody reads eight rows of choices. The databases
 * are called out because they are the ones certain to break - the image reads
 * its password only when it initializes, so it keeps the one the old data was
 * created with while the app is given a freshly generated one.
 */
export function StorageInUseDialog({ open, findings, issues, isPending, onOpenChange, onConfirm }: Props) {
    const databases = findings.filter(finding => finding.isDatabase);
    // Creating would fail on these, so the buttons that create are not offered:
    // the address or the port has to be changed first.
    const blocked = issues.length > 0;

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
                    <DialogTitle>{blocked ? "This cannot be created yet" : "These apps already have data"}</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <DialogBody className="flex flex-col gap-4">
                    {blocked && (
                        <ul className="flex flex-col gap-2">
                            {issues.map(issue => (
                                <li
                                    key={issue.code + issue.detail}
                                    className="flex items-start gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive"
                                >
                                    <AlertTriangle className="mt-0.5 size-4 shrink-0" />
                                    <span>{issue.detail}</span>
                                </li>
                            ))}
                        </ul>
                    )}

                    {findings.length > 0 && (
                        <>
                            <p className="text-sm leading-6">
                                {findings.length === 1 ? "One app" : `${findings.length} apps`} would be created on a
                                directory that a previous install left behind, and would start with what is in it.
                            </p>

                            <ul className="flex flex-col gap-1.5 rounded-md border border-border/80 bg-muted/30 px-3.5 py-2.5">
                                {findings.map(finding => (
                                    <li
                                        key={`${finding.appKey}:${finding.path}`}
                                        className="flex flex-wrap items-baseline gap-x-2 text-sm"
                                    >
                                        <span className="font-medium">{finding.app}</span>
                                        <span className="font-mono text-xs text-muted-foreground">
                                            {finding.volume.name}/{finding.path}
                                        </span>
                                        {finding.isDatabase && (
                                            <span className="text-xs font-medium text-destructive">database</span>
                                        )}
                                    </li>
                                ))}
                            </ul>

                            {databases.length > 0 && (
                                <div className="flex items-start gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive">
                                    <AlertTriangle className="mt-0.5 size-4 shrink-0" />
                                    <span>
                                        {databases.length === 1
                                            ? `${databases[0]?.app} is a database and will not start: `
                                            : "The databases above will not start: "}
                                        the password is generated fresh for this app, while the data on disk keeps the
                                        one it was created with.
                                    </span>
                                </div>
                            )}

                            {!blocked && (
                                <p className="text-xs text-muted-foreground">
                                    Deleting removes only those directories. The volume, and anything else on it, stays.
                                </p>
                            )}
                        </>
                    )}
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
                    {!blocked && (
                        <>
                            <Button
                                variant="outline"
                                disabled={isPending}
                                onClick={() => {
                                    onConfirm(false);
                                }}
                            >
                                Create anyway
                            </Button>
                            <Button
                                variant="destructive"
                                disabled={isPending}
                                isLoading={isPending}
                                onClick={() => {
                                    onConfirm(true);
                                }}
                            >
                                Delete that data and create
                            </Button>
                        </>
                    )}
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
