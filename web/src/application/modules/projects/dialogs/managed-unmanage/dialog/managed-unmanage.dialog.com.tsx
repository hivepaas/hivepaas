import { useState } from "react";

import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { AppManagedCommands } from "~/projects/data";
import { ConfirmDangerTargetBadge } from "~/projects/module-shared/components";

import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { Button, Field, Input, Separator } from "@/components/ui";

import { useManagedUnmanageDialogState } from "../hooks";

/**
 * Makes a managed database an ordinary app, once its name is typed: what the
 * engine made goes, and what the database needs stays.
 */
export function ManagedUnmanageDialog() {
    const navigate = useNavigate();
    const { state, props: dialogOptions, ...actions } = useManagedUnmanageDialogState();
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    const [typed, setTyped] = useState("");

    const open = state.mode === "open";
    const { target } = state;

    const { mutate: unmanage, isPending } = AppManagedCommands.useUnmanage({
        onSuccess: (_, request) => {
            toast.success(`${target?.appName ?? "The app"} is no longer managed`);
            setTyped("");
            actions.close();
            dialogOptions?.onSuccess?.();
            void navigate(
                ROUTE.projects.single.apps.single.configuration.general.$route(
                    request.projectID,
                    request.env,
                    request.appID,
                ),
            );
        },
        onError: error => {
            dialogOptions?.onError?.(error);
        },
    });

    function handleClose(): void {
        if (isPending) {
            return;
        }

        setTyped("");
        actions.close();
        dialogOptions?.onClose?.();
    }

    const isConfirmed = target !== null && typed === target.appName;

    return (
        <Dialog
            open={open}
            onOpenChange={isOpen => {
                if (!isOpen) {
                    handleClose();
                }
            }}
        >
            <DialogFixedContent className="sm:max-w-[600px]">
                <DialogHeader>
                    <DialogTitle>Stop managing the database</DialogTitle>
                </DialogHeader>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                {target && (
                    <form
                        className="min-h-0 flex flex-1 flex-col"
                        onSubmit={event => {
                            event.preventDefault();
                            if (!canWrite || !isConfirmed) {
                                return;
                            }
                            unmanage({ projectID: target.projectId, env: target.env, appID: target.appId });
                        }}
                    >
                        <DialogBody className="flex flex-col gap-4 text-sm">
                            <div className="flex flex-col gap-1.5">
                                <p className="font-medium">What goes</p>
                                <ul className="list-disc pl-5 text-muted-foreground space-y-0.5">
                                    <li>The nightly base backup and its retention.</li>
                                    <li>The archiver&apos;s health check, and its notifications.</li>
                                    <li>The restore test.</li>
                                    {target.hasVerifier && <li>The verifier app the restore test runs in.</li>}
                                    <li>Restore points, and restores from this page.</li>
                                    <li>The locks: every setting can be changed again.</li>
                                </ul>
                            </div>
                            <div className="flex flex-col gap-1.5">
                                <p className="font-medium">What stays</p>
                                <ul className="list-disc pl-5 text-muted-foreground space-y-0.5">
                                    <li>The database, its data, and its WAL archiving, which goes on where it did.</li>
                                    <li>The archive&apos;s base backups and WAL, for you to keep or delete.</li>
                                    <li>The scripts, secrets and settings the engine gave the app.</li>
                                </ul>
                            </div>
                            <p className="font-medium leading-6">
                                To confirm, type <ConfirmDangerTargetBadge text={target.appName} /> below.
                            </p>
                            <Field>
                                <Input
                                    aria-label="App name"
                                    value={typed}
                                    onChange={event => {
                                        setTyped(event.target.value);
                                    }}
                                    disabled={!canWrite || isPending}
                                />
                            </Field>
                        </DialogBody>
                        <DialogActionFooter>
                            <Button
                                type="submit"
                                variant="destructive"
                                disabled={!canWrite || isPending || !isConfirmed}
                                isLoading={isPending}
                                className="min-w-[120px]"
                            >
                                Stop managing
                            </Button>
                        </DialogActionFooter>
                    </form>
                )}
            </DialogFixedContent>
        </Dialog>
    );
}
