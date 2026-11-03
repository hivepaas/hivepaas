import { useState } from "react";

import { Button } from "@components/ui/button";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { AlertTriangle } from "lucide-react";
import { toast } from "sonner";
import { BackupSnapshotCommands } from "~/settings/data/commands";
import type { BackupSnapshot, BackupSnapshotScope } from "~/settings/domain";

import { Input } from "@/components/ui";

/** Deleting a snapshot, confirmed by typing its short ID. */
export function BackupSnapshotDeleteDialog({ scope, snapshot, onOpenChange }: Props) {
    const [typed, setTyped] = useState("");
    const { mutate: deleteOne, isPending } = BackupSnapshotCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Backup snapshot deleted");
            setTyped("");
            onOpenChange(false);
        },
    });

    return (
        <Dialog
            open={Boolean(snapshot)}
            onOpenChange={next => {
                if (!isPending) {
                    setTyped("");
                    onOpenChange(next);
                }
            }}
        >
            <DialogFixedContent className="sm:max-w-[520px]">
                <DialogHeader>
                    <DialogTitle>Delete Backup Snapshot</DialogTitle>
                </DialogHeader>
                {snapshot && (
                    <DialogBody className="flex flex-col gap-4">
                        <p className="text-sm leading-6 text-foreground">
                            The snapshot <span className="font-mono font-semibold">{snapshot.shortId}</span> is deleted
                            from the repository <span className="font-semibold">{snapshot.repo.name}</span>. Its data
                            cannot be restored afterwards.
                        </p>
                        <div className="flex items-center gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive font-medium">
                            <AlertTriangle className="size-4 shrink-0 text-destructive" />
                            <span>Type the snapshot&apos;s short ID to confirm.</span>
                        </div>
                        <Input
                            value={typed}
                            onChange={event => {
                                setTyped(event.target.value);
                            }}
                            placeholder={snapshot.shortId}
                            className="font-mono"
                        />
                    </DialogBody>
                )}
                <DialogActionFooter className="flex justify-end">
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
                        isLoading={isPending}
                        disabled={typed.trim() !== snapshot?.shortId}
                        onClick={() => {
                            if (snapshot) {
                                deleteOne({ scope, id: snapshot.id });
                            }
                        }}
                    >
                        Delete
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}

interface Props {
    scope: BackupSnapshotScope;
    snapshot: BackupSnapshot | null;
    onOpenChange: (open: boolean) => void;
}
