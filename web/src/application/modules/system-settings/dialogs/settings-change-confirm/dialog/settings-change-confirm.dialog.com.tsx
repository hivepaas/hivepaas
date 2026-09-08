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
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangleIcon, CheckCircle2Icon, ExternalLinkIcon, Loader2Icon, WifiOffIcon } from "lucide-react";
import { toast } from "sonner";
import {
    HivePaaSRoutingSettingsCommands,
    HivePaaSRoutingSettingsQueries,
    HivePaaSServiceSettingsCommands,
    HivePaaSServiceSettingsQueries,
} from "~/system-settings/data";
import { QK } from "~/system-settings/data/constants";
import type { HivePaaSRoutingDomain } from "~/system-settings/domain";

import { HttpException } from "@infrastructure/exceptions/http";

import { formatCountdown, useNow, useSettingsChangeConfirmDialogState } from "../hooks";
import type { SettingsChangeKind, SettingsChangeOutcome } from "../types";

/** How often the probe asks whether HivePaaS is still reachable. */
const PROBE_INTERVAL_MS = 3000;

/**
 * Failed probes before the dialog calls it a disconnection.
 *
 * One failure is a blip. Two in a row, six seconds apart, is worth saying
 * something about - but only once the change is supposed to be live. See phase.
 */
const PROBE_FAILURES_BEFORE_DISCONNECTED = 2;

const COPY: Record<SettingsChangeKind, { title: string; applied: string; settling: string }> = {
    routing: {
        title: "Confirm routing change",
        applied: "The change is live.",
        settling: "Waiting for the proxy to pick the change up",
    },
    service: {
        title: "Confirm proxy settings change",
        // Traefik is restarted by this change, so the honest thing to say is that
        // the dashboard is expected to be unreachable for part of the countdown.
        applied: "The change is being applied. HivePaaS may be unreachable while the proxy restarts.",
        settling: "Waiting for the proxy to restart",
    },
};

type Phase = "waiting" | "ready" | "disconnected" | "expired" | "resolved";

const fnPlaceholder = () => null;

/**
 * The domain a verification tab should be pointed at.
 *
 * The first enabled one, because that is the domain traefik answers on by
 * default. Disabled entries are skipped rather than indexed past: offering a link
 * that cannot resolve would look like the change broke something it did not.
 */
function findPrimaryDomain(domains: HivePaaSRoutingDomain[] | undefined): string | null {
    const primary = domains?.find(domain => domain.enabled && domain.domain !== "");

    return primary?.domain ?? null;
}

export function SettingsChangeConfirmDialog() {
    const { state, ...actions } = useSettingsChangeConfirmDialogState();
    const queryClient = useQueryClient();

    const open = state.mode === "open";
    const kind: SettingsChangeKind = state.mode === "open" ? state.kind : "routing";
    const pendingChange = state.mode === "open" ? state.pendingChange : null;
    const changeId = pendingChange?.changeId ?? "";

    const now = useNow(open);
    const [tooEarly, setTooEarly] = useState(false);

    // Both probes are declared because hooks cannot be called conditionally; only
    // the one matching the trial is enabled, so only it makes requests.
    const routingProbe = HivePaaSRoutingSettingsQueries.useProbe({
        enabled: open && kind === "routing",
        refetchInterval: open && kind === "routing" ? PROBE_INTERVAL_MS : false,
    });
    const serviceProbe = HivePaaSServiceSettingsQueries.useProbe({
        enabled: open && kind === "service",
        refetchInterval: open && kind === "service" ? PROBE_INTERVAL_MS : false,
    });
    const probe = kind === "routing" ? routingProbe : serviceProbe;

    const msUntilConfirmable = pendingChange ? pendingChange.confirmableFrom.getTime() - now : 0;
    const msUntilDeadline = pendingChange ? pendingChange.deadlineAt.getTime() - now : 0;

    // The trial is over the moment the server stops reporting it, whoever ended
    // it. Comparing ids rather than checking for null also covers a second change
    // having started somewhere else in the meantime.
    const serverPendingId = probe.data?.data.pendingChange?.changeId ?? null;
    const resolvedByServer = probe.isSuccess && open && serverPendingId !== changeId;

    const disconnected = probe.failureCount >= PROBE_FAILURES_BEFORE_DISCONNECTED;

    // Order matters, and "waiting" sitting above "disconnected" is the whole point
    // of it.
    //
    // Before confirmableFrom the server has not claimed the change is live yet. A
    // proxy settings change restarts traefik, so being unable to reach HivePaaS in
    // that window is what is supposed to happen - reporting it as a lockout would
    // raise a false alarm on every single proxy change, next to a button offering
    // to revert. After that moment the same failure means something real.
    //
    // Past the deadline the story is always "it ran out", even once the probe
    // confirms the trial is gone: that is the same event, and the vaguer "no longer
    // on trial" wording would replace an accurate answer with a guess.
    const phase: Phase =
        msUntilDeadline <= 0
            ? "expired"
            : resolvedByServer
              ? "resolved"
              : msUntilConfirmable > 0
                ? "waiting"
                : disconnected
                  ? "disconnected"
                  : "ready";

    const onConfirmError = (error: Error) => {
        // Not a failure - the new configuration is not serving yet. The countdown
        // is already showing when it will be worth retrying.
        if (error instanceof HttpException && error.code === "ERR_SETTINGS_CONFIRM_TOO_EARLY") {
            setTooEarly(true);
            return;
        }
        // Anything else means this is no longer the change on trial. The probe
        // will notice on its next tick and move the dialog to "resolved".
        void probe.refetch();
    };

    // Closing out a trial is one thing, done in one place, so no caller can end
    // one without recording it. The record is what stops the module reopening the
    // dialog from a query result that has not caught up yet.
    const finish = (outcome: SettingsChangeOutcome) => {
        actions.markResolved(changeId);
        actions.close();

        // The confirm and revert commands invalidate their own settings query
        // already; this covers the outcomes that ran no command at all - a trial
        // that expired, or one another session ended.
        if (outcome === "expired" || outcome === "superseded") {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.hivepaas.routing-settings.find-one"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.hivepaas.service-settings.find-one"]],
            });
        }
    };

    const onConfirmed = () => {
        toast.success("Change confirmed");
        finish("confirmed");
    };

    const onReverted = () => {
        toast.info("Change reverted");
        finish("reverted");
    };

    const routingConfirm = HivePaaSRoutingSettingsCommands.useConfirmChange({
        onSuccess: onConfirmed,
        onError: onConfirmError,
    });
    const serviceConfirm = HivePaaSServiceSettingsCommands.useConfirmChange({
        onSuccess: onConfirmed,
        onError: onConfirmError,
    });
    const routingRevert = HivePaaSRoutingSettingsCommands.useRevertChange({
        onSuccess: onReverted,
        onError: () => void probe.refetch(),
    });
    const serviceRevert = HivePaaSServiceSettingsCommands.useRevertChange({
        onSuccess: onReverted,
        onError: () => void probe.refetch(),
    });

    const confirm = kind === "routing" ? routingConfirm : serviceConfirm;
    const revert = kind === "routing" ? routingRevert : serviceRevert;

    // Read off the routing probe when there is one: it is the configuration that
    // is actually live, not whatever the form was showing. A proxy change does not
    // touch the domain list, so the host being used now is the right thing to
    // check there.
    const primaryDomain = findPrimaryDomain(routingProbe.data?.data.domains) ?? window.location.hostname;
    const verificationUrl = `${window.location.protocol}//${primaryDomain}`;

    // Session cookies are scoped to a host, so a verification tab on a domain
    // other than this one lands on the login screen. That is the check passing -
    // the request reached HivePaaS through the new configuration - but it reads
    // like a failure to somebody who is already worried, and the button next to it
    // says "Revert now".
    const verificationNeedsSignIn = primaryDomain !== window.location.hostname;

    // Clear the "too early" notice as soon as it stops being true, so a retry that
    // is now allowed does not look like it is still blocked.
    useEffect(() => {
        if (msUntilConfirmable <= 0) {
            setTooEarly(false);
        }
    }, [msUntilConfirmable]);

    const isBusy = confirm.isPending || revert.isPending;
    const copy = COPY[kind];

    function handleResolvedAcknowledged() {
        finish(phase === "expired" ? "expired" : "superseded");
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
                    <DialogTitle>{copy.title}</DialogTitle>
                </DialogHeader>
                <DialogDescription className="sr-only">
                    The change has been applied and will be reverted unless it is confirmed.
                </DialogDescription>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>

                <DialogBody className="flex flex-col gap-4">
                    {(phase === "waiting" || phase === "ready" || phase === "disconnected") && (
                        <>
                            <p className="text-sm text-foreground">
                                {copy.applied} It will be <span className="font-semibold">reverted automatically</span>{" "}
                                unless you confirm it.
                            </p>

                            <div className="flex items-baseline justify-between rounded-lg border bg-background/50 p-3">
                                <span className="text-sm text-muted-foreground">Reverting in</span>
                                <span className="font-mono text-2xl font-semibold tabular-nums">
                                    {formatCountdown(msUntilDeadline)}
                                </span>
                            </div>

                            {phase === "waiting" && (
                                <p className="flex items-center gap-2 text-xs text-muted-foreground">
                                    <Loader2Icon className="size-3.5 shrink-0 animate-spin" />
                                    {copy.settling} ({formatCountdown(msUntilConfirmable)}). Confirming now would vouch
                                    for the previous configuration.
                                </p>
                            )}

                            {phase === "disconnected" && (
                                <div className="flex flex-col gap-2">
                                    <p className="flex items-center gap-2 text-sm font-semibold text-destructive">
                                        <WifiOffIcon className="size-4 shrink-0" />
                                        Cannot reach HivePaaS
                                    </p>
                                    <p className="text-sm text-foreground">
                                        This is what the countdown is for. HivePaaS will revert the change on its own at{" "}
                                        <span className="font-semibold">
                                            {pendingChange?.deadlineAt.toLocaleTimeString()}
                                        </span>{" "}
                                        from inside the cluster, where it does not need your connection.
                                    </p>
                                    <p className="text-sm font-semibold text-foreground">
                                        Nothing else is needed. Wait for the countdown, then reload the page.
                                    </p>
                                </div>
                            )}

                            {phase === "ready" && (
                                <div className="flex flex-col gap-2">
                                    <p className="flex items-center gap-2 text-xs text-emerald-600 dark:text-emerald-400">
                                        <CheckCircle2Icon className="size-3.5 shrink-0" />
                                        HivePaaS is still reachable through the new configuration.
                                    </p>
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
                                                That is a different host, so it will ask you to sign in. Reaching the
                                                login page is itself the proof: the domain, its certificate and the new
                                                routing all work.
                                            </p>
                                        )}
                                        <p className="text-xs text-muted-foreground">
                                            Leave this tab open - make sure you come back and confirm before the
                                            countdown runs out.
                                        </p>
                                    </div>
                                </div>
                            )}

                            {tooEarly && (
                                <p className="text-xs text-amber-600 dark:text-amber-400">
                                    Too early to confirm. Try again when the countdown above allows it.
                                </p>
                            )}
                        </>
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
                                isLoading={revert.isPending}
                                onClick={() => {
                                    revert.mutate({ changeId });
                                }}
                            >
                                Revert now
                            </Button>
                            <Button
                                type="button"
                                className="min-w-[100px]"
                                disabled={isBusy || phase !== "ready"}
                                isLoading={confirm.isPending}
                                onClick={() => {
                                    confirm.mutate({ changeId });
                                }}
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
