import { useEffect, useState } from "react";

import { Button } from "@components/ui/button";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogDescription,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { Separator } from "@components/ui/separator";
import { AlertTriangleIcon, CheckCircle2Icon, ExternalLinkIcon, Loader2Icon, WifiOffIcon } from "lucide-react";
import { toast } from "sonner";
import { HivePaaSRoutingSettingsCommands, HivePaaSRoutingSettingsQueries } from "~/system-settings/data";
import type { HivePaaSRoutingDomain } from "~/system-settings/domain";

import { HttpException } from "@infrastructure/exceptions/http";

import { formatCountdown, useNow, useRoutingChangeConfirmDialogState } from "../hooks";

/** How often the probe asks whether HivePaaS is still reachable. */
const PROBE_INTERVAL_MS = 3000;

/**
 * The domain a verification tab should be pointed at.
 *
 * The first enabled one, because that is the domain traefik answers on by
 * default and the one an operator means when they say "the dashboard". Disabled
 * entries are skipped rather than indexed past: offering a link that cannot
 * resolve would look like the change broke something it did not.
 */
function findPrimaryDomain(domains: HivePaaSRoutingDomain[] | undefined): string | null {
    const primary = domains?.find(domain => domain.enabled && domain.domain !== "");

    return primary?.domain ?? null;
}

/**
 * Failed probes before the dialog calls it a disconnection.
 *
 * One failure is a blip. Two in a row, six seconds apart, is the operator having
 * locked themselves out - and that is the moment they need to be told they do
 * not have to do anything about it.
 */
const PROBE_FAILURES_BEFORE_DISCONNECTED = 2;

type Phase = "waiting" | "ready" | "disconnected" | "expired" | "resolved";

const fnPlaceholder = () => null;

export function RoutingChangeConfirmDialog() {
    const { state, props: { onResolved = fnPlaceholder } = {}, ...actions } = useRoutingChangeConfirmDialogState();

    const open = state.mode === "open";
    const pendingChange = state.mode === "open" ? state.pendingChange : null;
    const changeId = pendingChange?.changeId ?? "";

    const now = useNow(open);
    const [tooEarly, setTooEarly] = useState(false);

    const probe = HivePaaSRoutingSettingsQueries.useProbe({
        enabled: open,
        refetchInterval: open ? PROBE_INTERVAL_MS : false,
    });

    const msUntilConfirmable = pendingChange ? pendingChange.confirmableFrom.getTime() - now : 0;
    const msUntilDeadline = pendingChange ? pendingChange.deadlineAt.getTime() - now : 0;

    // The trial is over the moment the server stops reporting it, whoever ended
    // it. Comparing ids rather than checking for null also covers a second change
    // having started somewhere else in the meantime.
    const serverPendingId = probe.data?.data.pendingChange?.changeId ?? null;
    const resolvedByServer = probe.isSuccess && open && serverPendingId !== changeId;

    const disconnected = probe.failureCount >= PROBE_FAILURES_BEFORE_DISCONNECTED;

    // Read off the probe, so it is the configuration that is actually live - not
    // whatever the form happened to be showing when Save was pressed.
    //
    // The protocol comes from this page rather than being assumed https: a dev
    // install served over plain http would otherwise be handed a link to a
    // certificate it does not have.
    const primaryDomain = findPrimaryDomain(probe.data?.data.domains);
    const verificationUrl = primaryDomain == null ? null : `${window.location.protocol}//${primaryDomain}`;

    // Session cookies are scoped to a host, so a verification tab on a domain
    // other than this one lands on the login screen. That is the check passing -
    // the request reached HivePaaS through the new routing - but it reads like a
    // failure to somebody who is already worried they have broken something, and
    // the button next to it says "Revert now".
    const verificationNeedsSignIn = primaryDomain != null && primaryDomain !== window.location.hostname;

    // Order matters. Past the deadline the story is always "it ran out", even once
    // the probe confirms the trial is gone - that is the same event, and the vaguer
    // "no longer on trial" wording would replace an accurate answer with a guess at
    // exactly the moment the truth became known.
    const phase: Phase =
        msUntilDeadline <= 0
            ? "expired"
            : resolvedByServer
              ? "resolved"
              : disconnected
                ? "disconnected"
                : msUntilConfirmable > 0
                  ? "waiting"
                  : "ready";

    const { mutate: confirmChange, isPending: isConfirming } = HivePaaSRoutingSettingsCommands.useConfirmChange({
        onSuccess: () => {
            toast.success("Routing change confirmed");
            onResolved("confirmed", changeId);
            actions.close();
        },
        onError: error => {
            // Not a failure - the proxy has not finished picking the change up.
            // The countdown is already showing when it will be worth retrying.
            if (error instanceof HttpException && error.code === "ERR_SETTINGS_CONFIRM_TOO_EARLY") {
                setTooEarly(true);
                return;
            }
            // Anything else means this is no longer the change on trial. The probe
            // will notice on its next tick and move the dialog to "resolved".
            void probe.refetch();
        },
    });

    const { mutate: revertChange, isPending: isReverting } = HivePaaSRoutingSettingsCommands.useRevertChange({
        onSuccess: () => {
            toast.info("Routing change reverted");
            onResolved("reverted", changeId);
            actions.close();
        },
        onError: () => {
            void probe.refetch();
        },
    });

    // Clear the "too early" notice as soon as it stops being true, so a retry
    // that is now allowed does not look like it is still blocked.
    useEffect(() => {
        if (msUntilConfirmable <= 0) {
            setTooEarly(false);
        }
    }, [msUntilConfirmable]);

    const isBusy = isConfirming || isReverting;

    function handleConfirm() {
        confirmChange({ changeId });
    }

    function handleRevert() {
        revertChange({ changeId });
    }

    function handleResolvedAcknowledged() {
        onResolved(phase === "expired" ? "expired" : "superseded", changeId);
        actions.close();
    }

    return (
        <Dialog
            open={open}
            // Deliberately not dismissible. Closing this is the one action that
            // silently gives up a change the operator just made, and clicking the
            // backdrop is not a way anybody means to do that.
            onOpenChange={fnPlaceholder}
        >
            <DialogFixedContent
                className="w-full sm:w-[520px] sm:max-w-[520px]"
                showCloseButton={false}
                onEscapeKeyDown={event => {
                    event.preventDefault();
                }}
                onInteractOutside={event => {
                    event.preventDefault();
                }}
            >
                <DialogHeader>
                    <DialogTitle>Confirm routing change</DialogTitle>
                </DialogHeader>
                <DialogDescription className="sr-only">
                    The routing change has been applied and will be reverted unless it is confirmed.
                </DialogDescription>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>

                <DialogBody className="flex flex-col gap-4">
                    {(phase === "waiting" || phase === "ready") && (
                        <>
                            <p className="text-sm text-foreground">
                                The change is live. It will be{" "}
                                <span className="font-semibold">reverted automatically</span> unless you confirm it.
                            </p>

                            <div className="flex items-baseline justify-between rounded-lg border bg-background/50 p-3">
                                <span className="text-sm text-muted-foreground">Reverting in</span>
                                <span className="font-mono text-2xl font-semibold tabular-nums">
                                    {formatCountdown(msUntilDeadline)}
                                </span>
                            </div>

                            {phase === "waiting" ? (
                                <p className="flex items-center gap-2 text-xs text-muted-foreground">
                                    <Loader2Icon className="size-3.5 shrink-0 animate-spin" />
                                    Waiting for the proxy to pick the change up ({formatCountdown(msUntilConfirmable)}).
                                    Confirming now would vouch for the previous configuration.
                                </p>
                            ) : probe.isSuccess ? (
                                <div className="flex flex-col gap-2">
                                    <p className="flex items-center gap-2 text-xs text-emerald-600 dark:text-emerald-400">
                                        <CheckCircle2Icon className="size-3.5 shrink-0" />
                                        HivePaaS is still reachable through the new configuration.
                                    </p>
                                    {verificationUrl != null && (
                                        <div className="flex flex-col gap-1">
                                            <p className="text-xs text-muted-foreground">
                                                Want to check it further?{" "}
                                                <a
                                                    href={verificationUrl}
                                                    target="_blank"
                                                    rel="noopener noreferrer"
                                                    className="text-link inline-flex items-center gap-1 underline-offset-4 hover:underline"
                                                >
                                                    Open {primaryDomain} in a new tab
                                                    <ExternalLinkIcon className="size-3 shrink-0" />
                                                </a>
                                            </p>
                                            {verificationNeedsSignIn && (
                                                <p className="text-xs text-muted-foreground">
                                                    That is a different host, so it will ask you to sign in. Reaching
                                                    the login page is itself the proof: the domain, its certificate and
                                                    the new routing all work.
                                                </p>
                                            )}
                                            <p className="text-xs text-muted-foreground">
                                                Leave this tab open - make sure you come back and confirm before the
                                                countdown runs out.
                                            </p>
                                        </div>
                                    )}
                                </div>
                            ) : (
                                <p className="flex items-center gap-2 text-xs text-muted-foreground">
                                    <Loader2Icon className="size-3.5 shrink-0 animate-spin" />
                                    Checking the connection...
                                </p>
                            )}

                            {tooEarly && (
                                <p className="text-xs text-amber-600 dark:text-amber-400">
                                    Too early to confirm. Try again when the countdown above allows it.
                                </p>
                            )}
                        </>
                    )}

                    {phase === "disconnected" && (
                        <div className="flex flex-col gap-3">
                            <p className="flex items-center gap-2 text-sm font-semibold text-destructive">
                                <WifiOffIcon className="size-4 shrink-0" />
                                Cannot reach HivePaaS
                            </p>
                            <p className="text-sm text-foreground">
                                This is what the countdown is for. HivePaaS will revert the change on its own at{" "}
                                <span className="font-semibold">{pendingChange?.deadlineAt.toLocaleTimeString()}</span>{" "}
                                ({formatCountdown(msUntilDeadline)} from now), from inside the cluster where it does not
                                need your connection.
                            </p>
                            <p className="text-sm font-semibold text-foreground">
                                Nothing else is needed. Wait for the countdown, then reload the page.
                            </p>
                        </div>
                    )}

                    {phase === "expired" && (
                        <div className="flex flex-col gap-3">
                            <p className="flex items-center gap-2 text-sm font-semibold text-foreground">
                                <AlertTriangleIcon className="size-4 shrink-0 text-amber-500" />
                                The change was not confirmed in time
                            </p>
                            <p className="text-sm text-muted-foreground">
                                HivePaaS has reverted it to the previous configuration. Reload the page to see where the
                                settings stand.
                            </p>
                        </div>
                    )}

                    {phase === "resolved" && (
                        <div className="flex flex-col gap-3">
                            <p className="text-sm font-semibold text-foreground">This change is no longer on trial</p>
                            <p className="text-sm text-muted-foreground">
                                It was confirmed, reverted, or replaced by another change - possibly from a different
                                session. Reload the page to see where the settings stand.
                            </p>
                        </div>
                    )}
                </DialogBody>

                <DialogActionFooter>
                    {phase === "expired" || phase === "resolved" ? (
                        <Button
                            type="button"
                            className="min-w-[100px]"
                            onClick={handleResolvedAcknowledged}
                        >
                            Reload settings
                        </Button>
                    ) : (
                        <>
                            <Button
                                type="button"
                                variant="outline"
                                className="min-w-[100px]"
                                disabled={isBusy || phase === "disconnected"}
                                isLoading={isReverting}
                                onClick={handleRevert}
                            >
                                Revert now
                            </Button>
                            <Button
                                type="button"
                                className="min-w-[100px]"
                                disabled={isBusy || phase !== "ready"}
                                isLoading={isConfirming}
                                onClick={handleConfirm}
                            >
                                Keep change
                            </Button>
                        </>
                    )}
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
