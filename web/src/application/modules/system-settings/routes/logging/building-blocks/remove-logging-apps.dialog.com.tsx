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

/** What the operator types to confirm. */
const CONFIRM_TEXT = "logging";

interface Props {
    open: boolean;
    /**
     * Switching logging off takes both apps down; handing the backend to a store
     * somebody else runs takes only the backend down. The collector stays and
     * writes to the new store.
     */
    reason: "switch-off" | "external-backend";
    /** Whether HivePaaS's backend is one of the apps going away, with its stored logs. */
    removesBackend: boolean;
    isPending: boolean;
    onOpenChange: (open: boolean) => void;
    onConfirm: (removeStorage: boolean) => void;
}

/**
 * The logging stack is two apps HivePaaS runs in its hidden project, which is left
 * out of every listing. So the confirmation that would have lived on the apps' own
 * screens lives here, the way the registry's does.
 */
export function RemoveLoggingAppsDialog({ open, reason, removesBackend, isPending, onOpenChange, onConfirm }: Props) {
    const [typed, setTyped] = useState("");
    const [removeStorage, setRemoveStorage] = useState(false);

    useEffect(() => {
        if (open) {
            setTyped("");
            setRemoveStorage(false);
        }
    }, [open]);

    const confirmed = typed.trim() === CONFIRM_TEXT;
    const switchingOff = reason === "switch-off";

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
                    <DialogTitle>{switchingOff ? "Switch logging off" : "Use another log store"}</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <DialogBody className="flex flex-col gap-4">
                    <p className="text-sm leading-6">
                        {switchingOff
                            ? "This removes the apps that collect and store the logs, and the services they run. " +
                              "Logs stop being collected, and stored logs stop being shown in the apps' screens."
                            : "This removes the app that stores the logs, and the service it runs. The collector " +
                              "stays, and writes to the store you gave instead."}
                    </p>

                    <div className="flex items-center gap-2.5 rounded-md border border-destructive bg-destructive/10 px-3.5 py-2.5 text-sm font-medium text-destructive">
                        <AlertTriangle className="size-4 shrink-0" />
                        <span>Warning: this cannot be undone.</span>
                    </div>

                    {removesBackend && (
                        <label
                            className="flex items-start gap-2.5 text-sm leading-6"
                            htmlFor="logging-remove-storage"
                        >
                            <Checkbox
                                id="logging-remove-storage"
                                className="mt-1"
                                checked={removeStorage}
                                disabled={isPending}
                                onCheckedChange={checked => {
                                    setRemoveStorage(checked === true);
                                }}
                            />
                            <span>
                                <span className="block">Also delete the stored logs</span>
                                <span className="block text-xs leading-normal text-muted-foreground">
                                    The backend&apos;s own directory on the volume. The volume itself stays.
                                </span>
                            </span>
                        </label>
                    )}

                    <div className="flex flex-col gap-2">
                        <p className="text-sm font-medium leading-6">
                            Type <ConfirmDangerTargetBadge text={CONFIRM_TEXT} /> to confirm.
                        </p>
                        <Input
                            value={typed}
                            disabled={isPending}
                            placeholder={CONFIRM_TEXT}
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
                        {switchingOff ? "Switch off and remove" : "Remove the backend"}
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
