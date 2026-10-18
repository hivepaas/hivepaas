import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { History, KeyRound, Loader2, RefreshCw, Unplug } from "lucide-react";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppManagedCommands, AppManagedQueries, ProjectAppsQueries } from "~/projects/data";
import { useManagedRestoreDialog } from "~/projects/dialogs/managed-restore";
import { useManagedUnmanageDialog } from "~/projects/dialogs/managed-unmanage";
import type { AppManagedRestoreTarget, AppManagedSettings } from "~/projects/domain";
import { formatManagedTime, isRestoring } from "~/projects/module-shared/utils";

import { AppLink, AppLoader, PopConfirm } from "@application/shared/components";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { Button } from "@/components/ui/button";

import {
    ManagedRestorePoints,
    ManagedSettingsForm,
    ManagedStatusCards,
    ManagedSwitch,
    ManagedSwitchedBanner,
} from "../building-blocks";

const STATUS_REFETCH_INTERVAL_MS = 30_000;
const RESTORING_REFETCH_INTERVAL_MS = 10_000;

/**
 * A managed database's engine, as a person sees and steers it: whether its WAL
 * is archived and how far back a restore can go; restores into a new app; the
 * restore points; the settings it lets a person change; and the way out of
 * being managed.
 */
export function AppDatabaseRoute() {
    const { id: projectID, env, appId: appID } = useParams<{ id: string; env: string; appId: string }>();
    invariant(projectID, "projectID must be defined");
    invariant(env, "env must be defined");
    invariant(appID, "appID must be defined");
    const scope = { projectID, env, appID };

    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });
    // The header's query: the app is already loaded.
    const { data: app } = ProjectAppsQueries.useFindOneById({ projectID, env, appID, getStats: true });
    const appName = app?.data.name ?? "";

    const {
        data: settingsData,
        isLoading: isLoadingSettings,
        error: settingsError,
    } = AppManagedQueries.useFindSettings(scope, {
        // A restored database is watched until it has recovered.
        refetchInterval: query =>
            query.state.data && isRestoring(query.state.data.data) ? RESTORING_REFETCH_INTERVAL_MS : false,
    });
    const settings = settingsData?.data;
    const {
        data: statusData,
        isLoading: isLoadingStatus,
        isFetching: isFetchingStatus,
        refetch: refetchStatus,
    } = AppManagedQueries.useFindStatus(scope, {
        enabled: Boolean(settings),
        refetchInterval: STATUS_REFETCH_INTERVAL_MS,
    });
    const status = statusData?.data;

    const { actions: restoreDialog } = useManagedRestoreDialog();
    const { actions: unmanageDialog } = useManagedUnmanageDialog();
    const { mutate: refreshCredentials, isPending: isRefreshing } = AppManagedCommands.useRefreshCredentials({
        onSuccess: response => {
            toast.success(
                response.data.changed
                    ? "Credentials refreshed: the database restarts with them"
                    : "The credentials were current: nothing changed",
            );
        },
    });

    if (isLoadingSettings) {
        return (
            <div className={listBox}>
                <AppLoader />
            </div>
        );
    }

    if (settingsError || !settings) {
        return (
            <div className={cn(listBox, "text-sm text-muted-foreground")}>
                No engine looks after this app: it is not a managed database.
            </div>
        );
    }

    const restoring = isRestoring(settings);

    return (
        <div className={cn(listBox, "@container flex flex-col gap-6")}>
            <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex flex-col">
                    <h1 className="text-base font-semibold">Managed database</h1>
                    <p className="text-xs text-muted-foreground">
                        {engineName(settings.engine)}, archived continuously
                        {settings.storage?.prefix ? ` under ${settings.storage.prefix}` : ""}.
                    </p>
                </div>
                <div className="flex items-center gap-2">
                    <Button
                        variant="outline"
                        size="icon-sm"
                        aria-label="Refresh status"
                        title="Refresh status"
                        onClick={() => {
                            void refetchStatus();
                        }}
                    >
                        <RefreshCw className={cn("size-3.5", isFetchingStatus && "animate-spin")} />
                    </Button>
                    <Button
                        size="sm"
                        disabled={!canWrite || restoring || !status?.restoreWindow}
                        title={status?.restoreWindow ? undefined : "There is no base backup to restore from yet"}
                        onClick={() => {
                            restoreDialog.open({ projectId: projectID, env, appId: appID, appName });
                        }}
                    >
                        <History className="size-3.5" />
                        Restore
                    </Button>
                </div>
            </div>

            {restoring && settings.restore && (
                <RestoringBanner
                    projectId={projectID}
                    env={env}
                    appId={appID}
                    restore={settings.restore}
                />
            )}

            {settings.switchedTo && (
                <ManagedSwitchedBanner
                    projectId={projectID}
                    env={env}
                    switchedTo={settings.switchedTo}
                />
            )}

            {isLoadingStatus || !status ? (
                <AppLoader />
            ) : (
                <>
                    <ManagedStatusCards
                        settings={settings}
                        status={status}
                        snapshotsHref={ROUTE.projects.single.apps.single.configuration.backupSnapshots.$route(
                            projectID,
                            env,
                            appID,
                        )}
                    />
                    <ManagedRestorePoints
                        projectId={projectID}
                        env={env}
                        appId={appID}
                        appName={appName}
                        points={status.restorePoints}
                        canWrite={canWrite}
                        canMark={!restoring}
                    />
                </>
            )}

            {settings.restore?.readyAt && (
                <ManagedSwitch
                    projectId={projectID}
                    env={env}
                    appId={appID}
                    appName={appName}
                    canWrite={canWrite}
                />
            )}

            <ManagedSettingsForm
                key={settings.updateVer}
                projectId={projectID}
                env={env}
                appId={appID}
                settings={settings}
                readOnly={!canWrite}
            />

            <section
                aria-label="Maintenance"
                className="flex flex-col gap-3"
            >
                <h2 className="text-sm font-semibold">Maintenance</h2>
                <div className="flex flex-col divide-y rounded-md border">
                    <div className="flex flex-wrap items-center justify-between gap-3 px-3 py-3">
                        <div className="flex max-w-[640px] flex-col gap-0.5">
                            <span className="text-sm font-medium">Refresh credentials</span>
                            <span className="text-xs text-muted-foreground">
                                The archive&apos;s credentials were copied into the app when it was made. After a key is
                                rotated on the storage, this copies them again; the database restarts if they changed.
                            </span>
                        </div>
                        <PopConfirm
                            title="Refresh the archive's credentials?"
                            description="The database restarts if they changed."
                            confirmText="Refresh"
                            onConfirm={() => {
                                refreshCredentials(scope);
                            }}
                        >
                            <Button
                                variant="outline"
                                size="sm"
                                disabled={!canWrite || isRefreshing}
                            >
                                {isRefreshing ? (
                                    <Loader2 className="size-3.5 animate-spin" />
                                ) : (
                                    <KeyRound className="size-3.5" />
                                )}
                                Refresh credentials
                            </Button>
                        </PopConfirm>
                    </div>
                    <div className="flex flex-wrap items-center justify-between gap-3 px-3 py-3">
                        <div className="flex max-w-[640px] flex-col gap-0.5">
                            <span className="text-sm font-medium">Stop managing</span>
                            <span className="text-xs text-muted-foreground">
                                Makes it an ordinary app: the engine&apos;s jobs and locks go, the database and its
                                archive stay.
                            </span>
                        </div>
                        <Button
                            variant="destructiveOutline"
                            size="sm"
                            disabled={!canWrite}
                            onClick={() => {
                                unmanageDialog.open({
                                    projectId: projectID,
                                    env,
                                    appId: appID,
                                    appName,
                                    hasVerifier: Boolean(settings.restoreTest?.verifierApp),
                                });
                            }}
                        >
                            <Unplug className="size-3.5" />
                            Stop managing
                        </Button>
                    </div>
                </div>
            </section>
        </div>
    );
}

function RestoringBanner({
    projectId,
    env,
    appId,
    restore,
}: {
    projectId: string;
    env: string;
    appId: string;
    restore: NonNullable<AppManagedSettings["restore"]>;
}) {
    const sourceId = restore.sourceApp?.id;

    return (
        <div
            role="status"
            className="flex items-start gap-2 rounded-md border border-purple-500/30 bg-purple-500/10 px-3 py-2.5 text-xs text-purple-800 dark:text-purple-300"
        >
            <Loader2 className="mt-0.5 size-3.5 shrink-0 animate-spin" />
            <div className="flex flex-col gap-1">
                <span className="font-medium">
                    Restoring{restore.target ? ` to ${targetText(restore.target)}` : ""}
                    {sourceId && (
                        <>
                            {" "}
                            from{" "}
                            <AppLink.Basic
                                to={ROUTE.projects.single.apps.single.database.$route(projectId, env, sourceId)}
                                className="underline"
                            >
                                its source
                            </AppLink.Basic>
                        </>
                    )}
                </span>
                <span>
                    Its base backups, health check and restore test start once it has recovered and made its first base
                    backup. The restore runs in a{" "}
                    <AppLink.Basic
                        to={ROUTE.projects.single.apps.single.tasks.$route(projectId, env, appId)}
                        className="underline"
                    >
                        task
                    </AppLink.Basic>
                    .
                </span>
            </div>
        </div>
    );
}

function targetText(target: AppManagedRestoreTarget): string {
    switch (target.kind) {
        case "time":
            return target.time ? formatManagedTime(target.time) : "a time";
        case "point":
            return `the point ${target.point}`;
        case "backup":
            return `the base backup ${target.backup}`;
        default:
            return "the latest archived";
    }
}

function engineName(engine: string): string {
    return engine === "postgres" ? "PostgreSQL" : engine;
}
