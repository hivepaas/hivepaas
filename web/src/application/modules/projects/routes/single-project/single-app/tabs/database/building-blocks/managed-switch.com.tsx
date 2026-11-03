import { useState } from "react";

import { ArrowRightLeft, Loader2, PauseCircle } from "lucide-react";
import { toast } from "sonner";
import { AppManagedCommands, AppManagedQueries } from "~/projects/data";
import type { AppManagedSettings, AppManagedSwitchConsumer } from "~/projects/domain";
import { formatManagedTime } from "~/projects/module-shared/utils";

import { AppLink, PopConfirm } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Button, Checkbox } from "@/components/ui";

interface Props {
    projectId: string;
    env: string;
    appId: string;
    appName: string;
    canWrite: boolean;
}

/**
 * On a restored database, once it is ready: the apps that use the database it
 * was restored from, to switch to this one. The source keeps running, its jobs
 * paused, for the person to stop or delete once the switch is checked.
 */
export function ManagedSwitch({ projectId, env, appId, appName, canWrite }: Props) {
    const scope = { projectID: projectId, env, appID: appId };
    const { data, isLoading } = AppManagedQueries.useFindSwitch(scope);
    const info = data?.data;
    // The consumers left out; every one found is switched unless unchecked.
    const [excluded, setExcluded] = useState<Set<string>>(new Set());
    const [redeploy, setRedeploy] = useState(false);

    const { mutate: switchConsumers, isPending } = AppManagedCommands.useSwitch({
        onSuccess: response => {
            const { switched, deployments } = response.data;
            toast.success(
                `${switched.length} ${switched.length === 1 ? "app" : "apps"} switched to ${appName}` +
                    (deployments.length > 0 ? `, ${deployments.length} redeploying` : ""),
            );
            if (response.meta?.warning) {
                toast.warning(response.meta.warning);
            }
            setExcluded(new Set());
        },
    });

    if (isLoading || !info) {
        return null;
    }

    const { consumers } = info;
    const chosen = consumers.filter(consumer => !excluded.has(consumer.app.id));
    const sourceLink = (
        <AppLink.Basic
            to={ROUTE.projects.single.apps.single.database.$route(projectId, env, info.source.id)}
            className="underline"
        >
            {info.source.name}
        </AppLink.Basic>
    );

    return (
        <section
            aria-label="Switch consumers"
            className="flex flex-col gap-3"
        >
            <div className="flex flex-col">
                <h2 className="text-sm font-semibold">Switch consumers</h2>
                <p className="text-xs text-muted-foreground">
                    The apps that use {sourceLink}, which this database was restored from, name this one instead: their
                    variables, the databases their previews clone, and where their deployments mark a restore point.{" "}
                    {info.source.name} keeps running with its jobs paused: stop or delete it once the switch is checked.
                </p>
                {info.switchedAt && (
                    <p
                        className="text-xs text-muted-foreground"
                        title={formatManagedTime(info.switchedAt)}
                    >
                        Last switched {timeAgoFormatter.format(info.switchedAt)}.
                    </p>
                )}
            </div>
            <div className="rounded-md border px-3 py-3">
                {consumers.length === 0 ? (
                    <p className="text-xs text-muted-foreground">No app of this environment uses {info.source.name}.</p>
                ) : (
                    <div className="flex flex-col gap-3">
                        <ul className="flex flex-col gap-2">
                            {consumers.map(consumer => (
                                <li key={consumer.app.id}>
                                    <label
                                        htmlFor={`switch-consumer-${consumer.app.id}`}
                                        className="flex cursor-pointer select-none items-start gap-2 text-sm"
                                    >
                                        <Checkbox
                                            id={`switch-consumer-${consumer.app.id}`}
                                            aria-label={consumer.app.name}
                                            checked={!excluded.has(consumer.app.id)}
                                            disabled={!canWrite || isPending}
                                            onCheckedChange={checked => {
                                                setExcluded(current => {
                                                    const next = new Set(current);
                                                    if (checked === true) {
                                                        next.delete(consumer.app.id);
                                                    } else {
                                                        next.add(consumer.app.id);
                                                    }
                                                    return next;
                                                });
                                            }}
                                        />
                                        <span className="flex flex-col">
                                            <span className="font-medium">{consumer.app.name}</span>
                                            <span className="text-xs text-muted-foreground">{usesText(consumer)}</span>
                                        </span>
                                    </label>
                                </li>
                            ))}
                        </ul>
                        <label
                            htmlFor="switch-redeploy"
                            className="flex cursor-pointer select-none items-start gap-2 text-sm"
                        >
                            <Checkbox
                                id="switch-redeploy"
                                aria-label="Redeploy them"
                                checked={redeploy}
                                disabled={!canWrite || isPending}
                                onCheckedChange={checked => {
                                    setRedeploy(checked === true);
                                }}
                            />
                            <span className="flex flex-col">
                                <span>Redeploy them</span>
                                <span className="text-xs text-muted-foreground">
                                    Their variables are applied to their running services either way, which restart. A
                                    deployment also runs what reads the database as it builds or before it deploys.
                                </span>
                            </span>
                        </label>
                        <div className="flex justify-end">
                            <PopConfirm
                                title={`Switch ${chosen.length} ${chosen.length === 1 ? "app" : "apps"} to ${appName}?`}
                                description={`${info.source.name} keeps running, its jobs paused.`}
                                confirmText="Switch"
                                onConfirm={() => {
                                    switchConsumers({
                                        ...scope,
                                        payload: {
                                            consumers: chosen.map(consumer => ({ id: consumer.app.id })),
                                            redeploy,
                                        },
                                    });
                                }}
                            >
                                <Button
                                    size="sm"
                                    disabled={!canWrite || isPending || chosen.length === 0}
                                >
                                    {isPending ? (
                                        <Loader2 className="size-3.5 animate-spin" />
                                    ) : (
                                        <ArrowRightLeft className="size-3.5" />
                                    )}
                                    Switch
                                </Button>
                            </PopConfirm>
                        </div>
                    </div>
                )}
            </div>
        </section>
    );
}

function usesText(consumer: AppManagedSwitchConsumer): string {
    const uses: string[] = [];
    if (consumer.envVars.length > 0) {
        uses.push(`${consumer.envVars.length === 1 ? "variable" : "variables"} ${consumer.envVars.join(", ")}`);
    }
    if (consumer.previewClone) {
        uses.push("its previews clone it");
    }
    if (consumer.deployRestorePoint) {
        uses.push("its deployments mark a restore point on it");
    }
    return uses.join("; ");
}

/** On a database whose consumers were switched to one restored from it: its jobs are paused. */
export function ManagedSwitchedBanner({
    projectId,
    env,
    switchedTo,
}: {
    projectId: string;
    env: string;
    switchedTo: NonNullable<AppManagedSettings["switchedTo"]>;
}) {
    const targetId = switchedTo.app?.id;

    return (
        <div
            role="status"
            className="flex items-start gap-2 rounded-md border border-amber-500/30 bg-amber-500/10 px-3 py-2.5 text-xs text-amber-800 dark:text-amber-300"
        >
            <PauseCircle className="mt-0.5 size-3.5 shrink-0" />
            <div className="flex flex-col gap-1">
                <span className="font-medium">
                    Its consumers were switched to{" "}
                    {targetId ? (
                        <AppLink.Basic
                            to={ROUTE.projects.single.apps.single.database.$route(projectId, env, targetId)}
                            className="underline"
                        >
                            the database restored from it
                        </AppLink.Basic>
                    ) : (
                        "the database restored from it"
                    )}{" "}
                    <span title={formatManagedTime(switchedTo.at)}>{timeAgoFormatter.format(switchedTo.at)}</span>
                </span>
                <span>
                    Its base backups, health check, restore test and dump are paused. It keeps running, and archiving:
                    stop or delete it once the switch is checked.
                </span>
            </div>
        </div>
    );
}
