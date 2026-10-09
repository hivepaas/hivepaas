import { useEffect, useState } from "react";

import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { AlertTriangle, UserCheck, Users } from "lucide-react";
import type { AppStorageOwner } from "~/projects/api/services";
import { type OptionCard, OptionCardGroup } from "~/projects/module-shared/components/option-card-group";

import { Button, Label, Separator } from "@/components/ui";
import { InputNumber } from "@/components/ui/input-number";

type ResetMethod = "owner" | "everyone";

const METHOD_OPTIONS: OptionCard<ResetMethod>[] = [
    {
        value: "owner",
        label: "Give to a user",
        description: "Changes who owns the files and keeps their modes. A private key stays private.",
        icon: UserCheck,
    },
    {
        value: "everyone",
        label: "Open to every user",
        description: "Lets any user read and write them. For when the app's user is not known.",
        icon: Users,
    },
];

interface Props {
    open: boolean;
    /** The mount's path inside the container, to say which one this is. */
    target: string;
    isPending: boolean;
    onOpenChange: (open: boolean) => void;
    /** Without an owner, every user is let read and write the files instead. */
    onConfirm: (owner: AppStorageOwner | undefined) => void;
}

// parseId is the user or group id typed, or null while it is not one.
function parseId(text: string): number | null {
    const trimmed = text.trim();
    return /^\d+$/.test(trimmed) ? Number(trimmed) : null;
}

/**
 * Resets what one mount holds, for an app that has to be given data another user
 * wrote - a different image on the same volume, or files copied in by hand.
 *
 * Nothing does this on its own: a deployment opens a directory up only while it
 * is still empty, so what an app has written keeps the owner and modes it gave it.
 */
export function ResetPermissionsDialog({ open, target, isPending, onOpenChange, onConfirm }: Props) {
    const [method, setMethod] = useState<ResetMethod>("owner");
    // What is typed, as it is typed. A number field takes its number only once
    // it is left, and the button - still disabled when the second number is
    // typed - would take the click that leaves it, and do nothing.
    const [uidText, setUidText] = useState("");
    const [gidText, setGidText] = useState("");

    useEffect(() => {
        if (open) {
            setMethod("owner");
            setUidText("");
            setGidText("");
        }
    }, [open]);

    const uid = parseId(uidText);
    const gid = parseId(gidText);
    const ownerComplete = uid !== null && gid !== null;
    const canConfirm = method === "everyone" || ownerComplete;

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
            <DialogFixedContent className="sm:max-w-[600px]">
                <DialogHeader>
                    <DialogTitle>Reset permissions</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <DialogBody className="flex flex-col gap-4">
                    <p className="text-sm leading-6">
                        Changes the directory mounted at <span className="font-mono text-xs">{target}</span> and
                        everything in it. Use it when the app cannot read or write data another user wrote there - a
                        different image on the same volume, or files copied in by hand.
                    </p>

                    <OptionCardGroup
                        options={METHOD_OPTIONS}
                        value={method}
                        onChange={setMethod}
                        readOnly={isPending}
                        className="grid-cols-1 sm:grid-cols-2"
                    />

                    {method === "owner" ? (
                        <div className="flex flex-col gap-2">
                            <div className="flex gap-3">
                                <div className="flex flex-col gap-1.5">
                                    <Label htmlFor="reset-permissions-uid">User ID</Label>
                                    <InputNumber
                                        id="reset-permissions-uid"
                                        placeholder="1000"
                                        min={0}
                                        step={1}
                                        showControls={false}
                                        useGrouping={false}
                                        disabled={isPending}
                                        className="w-[140px]"
                                        onInput={event => {
                                            setUidText(event.currentTarget.value);
                                        }}
                                        onValueChange={value => {
                                            setUidText(value === undefined ? "" : String(value));
                                        }}
                                    />
                                </div>
                                <div className="flex flex-col gap-1.5">
                                    <Label htmlFor="reset-permissions-gid">Group ID</Label>
                                    <InputNumber
                                        id="reset-permissions-gid"
                                        placeholder="1000"
                                        min={0}
                                        step={1}
                                        showControls={false}
                                        useGrouping={false}
                                        disabled={isPending}
                                        className="w-[140px]"
                                        onInput={event => {
                                            setGidText(event.currentTarget.value);
                                        }}
                                        onValueChange={value => {
                                            setGidText(value === undefined ? "" : String(value));
                                        }}
                                    />
                                </div>
                            </div>
                            <p className="text-xs leading-5 text-muted-foreground">
                                The numbers the app writes its files as. Running <span className="font-mono">id</span>{" "}
                                in the app&apos;s terminal shows them - unless the image starts as root and switches
                                user itself, as databases do; then its documentation says which.
                            </p>
                        </div>
                    ) : (
                        <div className="flex items-start gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm text-destructive">
                            <AlertTriangle className="mt-0.5 size-4 shrink-0" />
                            <span>
                                Every file becomes readable and writable by anyone who can reach the volume, private
                                keys included. Some apps refuse to start on keys or directories other users can read.
                            </span>
                        </div>
                    )}

                    <p className="text-xs text-muted-foreground">
                        Symbolic links are left as they are and not followed. The app keeps running; restart it if it
                        failed to start.
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
                        disabled={!canConfirm || isPending}
                        isLoading={isPending}
                        onClick={() => {
                            onConfirm(method === "owner" && ownerComplete ? { uid, gid } : undefined);
                        }}
                    >
                        Reset permissions
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
