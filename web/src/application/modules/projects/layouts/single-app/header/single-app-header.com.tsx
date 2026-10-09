import { memo, useEffect, useMemo, useRef, useState } from "react";

import { Button, Checkbox } from "@components/ui";
import { Avatar } from "@components/ui/avatar";
import { useQueryClient } from "@tanstack/react-query";
import { Power, RefreshCw } from "lucide-react";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import {
    AppDeploymentsQueries,
    AppServiceTasksQueries,
    ProjectAppsCommands,
    ProjectAppsQueries,
    ProjectsQueries,
} from "~/projects/data";
import { QK } from "~/projects/data/constants";
import {
    AppActiveDeploymentBadge,
    AppActiveDeploymentDot,
    AppInstancesCountBadge,
    ProjectAppStatusBadge,
    ProjectEnvFilter,
} from "~/projects/module-shared/components";
import { EProjectAppStatus } from "~/projects/module-shared/enums";
import {
    APP_SERVICE_TASKS_REFETCH_INTERVAL_MS,
    computeAppInstancesHealth,
    isFunctionApp,
} from "~/projects/module-shared/utils";

import { PopConfirm, TabNavigation } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Separator } from "@/components/ui/separator";

import { SingleAppBreadcrumbs } from "../buidling-blocks";

import { AppAccessLinksDropdown } from "./building-blocks";
import { SingleAppHeaderSkeleton } from "./single-app-header.skeleton.com";

// How often the header asks whether a deployment of the app is queued or running.
const APP_ACTIVE_DEPLOYMENT_REFETCH_INTERVAL_MS = 5_000;

// useReadAgainWhenDeploymentEnds reads again what a deployment changes - the
// deployments, the app - once the one the header showed is no longer queued or
// running: it ended, or the next took its place.
function useReadAgainWhenDeploymentEnds(projectID: string, activeDeploymentId: string | null) {
    const queryClient = useQueryClient();
    const shown = useRef<string | null>(null);

    useEffect(() => {
        if (shown.current && shown.current !== activeDeploymentId) {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.apps.deployments.$.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.apps.deployments.$.find-one-by-id"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.apps.$.find-one-by-id"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.$.find-one-by-id"], { projectID }] });
        }
        shown.current = activeDeploymentId;
    }, [queryClient, projectID, activeDeploymentId]);
}

function View({ projectId, env, appId }: Props) {
    const { taskId } = useParams<{
        taskId?: string;
    }>();
    const { data, isLoading, error } = ProjectsQueries.useFindOneById({ projectID: projectId });
    const {
        data: app,
        isLoading: isLoadingApp,
        error: errorApp,
    } = ProjectAppsQueries.useFindOneById({ projectID: projectId, env, appID: appId, getStats: true });
    // Mounted for the whole app detail scope, so polling starts on entering the app and stops on leaving it.
    const { data: serviceTasksResponse, isSuccess: isServiceTasksLoaded } = AppServiceTasksQueries.useFindMany(
        { projectID: projectId, env, appID: appId },
        { refetchInterval: APP_SERVICE_TASKS_REFETCH_INTERVAL_MS },
    );
    const serviceTasks = serviceTasksResponse?.data;
    // A deployment queued or running, asked as often as the instances: the
    // header and the Deployments tab say so.
    const { data: activeDeploymentResponse } = AppDeploymentsQueries.useFindActive(
        { projectID: projectId, env, appID: appId },
        { refetchInterval: APP_ACTIVE_DEPLOYMENT_REFETCH_INTERVAL_MS },
    );
    const activeDeployment = activeDeploymentResponse?.data ?? null;
    useReadAgainWhenDeploymentEnds(projectId, activeDeployment?.id ?? null);
    const instancesHealth = useMemo(
        () => (serviceTasks ? computeAppInstancesHealth(serviceTasks) : null),
        [serviceTasks],
    );

    const [noCache, setNoCache] = useState(false);
    // The app a re-deploy was started for: its runtime notice goes, the deployment being on its way.
    const [redeployedAppId, setRedeployedAppId] = useState<string | null>(null);

    const { mutate: deploy, isPending: isDeploying } = ProjectAppsCommands.useDeploy({
        onSuccess: (_, request) => {
            toast.success("Re-deploy started");
            setRedeployedAppId(request.appID);
        },
    });
    const { mutate: restart, isPending: isRestarting } = ProjectAppsCommands.useRestart({
        onSuccess: () => {
            toast.success("Restart started");
        },
    });
    const { mutate: setRunning, isPending: isSettingRunning } = ProjectAppsCommands.useSetRunning({
        onSuccess: (_, request) => {
            toast.success(request.running ? "App start requested" : "App stop requested");
        },
    });
    const isAppActionPending = isDeploying || isRestarting || isSettingRunning;

    if (isLoading || isLoadingApp) {
        return <SingleAppHeaderSkeleton />;
    }

    if (error || errorApp) {
        return null;
    }

    invariant(data, "data must be defined");
    invariant(app, "app must be defined");
    const { data: project } = data;
    const { data: appData } = app;
    const isChildApp = Boolean(appData.parentApp);

    // "Stopped" means the service is scaled to 0, NOT that every task happens to be down.
    // An app whose replicas all crashed is broken, not stopped: it must offer Stop (and Re-deploy/
    // Restart), because Start on it would make the backend reset the replica count.
    // Prefer the polled task data; fall back to the app summary until the first poll lands.
    const desiredReplicas = instancesHealth?.total ?? appData.stats?.desiredTasks ?? 0;
    const isAppStopped = desiredReplicas === 0;
    const startStopText = isAppStopped ? "Start" : "Stop";

    // Nothing can be acted on while the app is being torn down. Other statuses stay actionable on
    // purpose - `missing` in particular is exactly when Re-deploy is needed to recreate the service.
    const isAppDeleting = appData.status === EProjectAppStatus.Deleting;
    const appEnv = project.envs.find(projectEnv => projectEnv.name === appData.env);
    const appRoute = ROUTE.projects.single.apps.single.configuration.general.$route(projectId, env, appId);
    const taskBreadcrumbItems = taskId
        ? [
              {
                  label: "Tasks",
                  to: ROUTE.projects.single.apps.single.tasks.$route(projectId, env, appId),
              },
              {
                  label: "Task Details",
              },
          ]
        : [];
    const configurationActivePathPrefixes = [
        ROUTE.projects.single.apps.single.configuration.deploymentSettings.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.routingSettings.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.envVariables.$route(projectId, env, appId),

        ROUTE.projects.single.apps.single.configuration.secrets.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.configFiles.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.settingMounts.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.containerSettings.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.availabilityAndScaling.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.persistentStorage.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.networks.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.resources.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.dockerApi.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.periodicJobs.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.scheduledJobs.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.appClone.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.featureSettings.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.dangerZone.$route(projectId, env, appId),
    ];

    const links = [
        ...(isFunctionApp(appData)
            ? [
                  {
                      route: ROUTE.projects.single.apps.single.code.$route(projectId, env, appId),
                      label: "Code",
                  },
              ]
            : []),
        {
            route: ROUTE.projects.single.apps.single.configuration.general.$route(projectId, env, appId),
            label: "Settings",
            activePathPrefixes: configurationActivePathPrefixes,
        },
        {
            route: ROUTE.projects.single.apps.single.instances.$route(projectId, env, appId),
            label: (
                <span className="inline-flex items-center gap-1.5">
                    Instances
                    {instancesHealth && isServiceTasksLoaded && (
                        <AppInstancesCountBadge
                            current={instancesHealth.current}
                            total={instancesHealth.total}
                        />
                    )}
                </span>
            ),
            activePathPrefixes: [ROUTE.projects.single.apps.single.instances.$route(projectId, env, appId)],
        },
        {
            route: ROUTE.projects.single.apps.single.deployments.$route(projectId, env, appId),
            label: (
                <span className="inline-flex items-center gap-1.5">
                    Deployments
                    {activeDeployment && <AppActiveDeploymentDot deployment={activeDeployment} />}
                </span>
            ),
            activePathPrefixes: [ROUTE.projects.single.apps.single.deployments.$route(projectId, env, appId)],
        },
        {
            route: ROUTE.projects.single.apps.single.tasks.$route(projectId, env, appId),
            label: "Tasks",
            activePathPrefixes: [ROUTE.projects.single.apps.single.tasks.$route(projectId, env, appId)],
        },
        // Every app's: its HTTP requests, and a function's calls.
        {
            route: ROUTE.projects.single.apps.single.metrics.$route(projectId, env, appId),
            label: "Metrics",
        },
        {
            route: ROUTE.projects.single.apps.single.logs.$route(projectId, env, appId),
            label: "Logs",
        },
        {
            route: ROUTE.projects.single.apps.single.terminal.$route(projectId, env, appId),
            label: "Terminal",
        },
        // A preview is built from a repository source: a function has none.
        ...(!isChildApp && !isFunctionApp(appData)
            ? [
                  {
                      route: ROUTE.projects.single.apps.single.previewDeployments.$route(projectId, env, appId),
                      label: "Preview Deployments",
                  },
              ]
            : []),
    ];
    return (
        <div className="bg-background pt-2 sm:pt-2.5 px-3 sm:px-5 rounded-lg">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 pb-2.5">
                <SingleAppBreadcrumbs
                    app={appData}
                    appRoute={appRoute}
                    items={taskBreadcrumbItems}
                    parentApp={appData.parentApp}
                    project={project}
                />
                {appEnv ? (
                    <div className="flex items-center">
                        <ProjectEnvFilter
                            projectId={projectId}
                            envs={[appEnv]}
                            showAll={false}
                            interactive={false}
                        />
                    </div>
                ) : null}
            </div>

            <Separator className="opacity-50" />

            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2.5 py-2.5">
                <div className="flex items-center gap-3 sm:gap-4 min-w-0">
                    <Avatar
                        name={appData.name}
                        src={appData.photo}
                        className="size-9 sm:size-12 text-sm sm:text-lg shrink-0 rounded-xl"
                    />
                    <div className="flex min-w-0 flex-col gap-1 justify-center">
                        <div className="flex flex-wrap items-center gap-2">
                            <h2 className="text-lg sm:text-xl font-semibold text-foreground truncate">
                                {appData.name}
                            </h2>
                            <ProjectAppStatusBadge status={appData.status} />
                            {activeDeployment && (
                                <AppActiveDeploymentBadge
                                    projectId={projectId}
                                    env={env}
                                    appId={appId}
                                    deployment={activeDeployment}
                                />
                            )}
                        </div>
                    </div>
                </div>

                <div className="flex flex-wrap items-center gap-2 sm:ml-auto">
                    <PopConfirm
                        title="Re-deploy app"
                        description="Are you sure you want to re-deploy this app?"
                        confirmText="Re-deploy"
                        cancelText="Cancel"
                        variant="destructive"
                        content={
                            <label
                                htmlFor="re-deploy-no-cache"
                                className="flex items-center gap-2 cursor-pointer select-none text-sm font-medium text-foreground"
                            >
                                <Checkbox
                                    id="re-deploy-no-cache"
                                    checked={noCache}
                                    onCheckedChange={checked => {
                                        setNoCache(checked === true);
                                    }}
                                />
                                <span>
                                    No cache <span className="text-xs text-muted-foreground font-normal">(SLOW)</span>
                                </span>
                            </label>
                        }
                        onOpenChange={open => {
                            if (!open) {
                                setNoCache(false);
                            }
                        }}
                        onConfirm={() => {
                            deploy({
                                projectID: projectId,
                                env,
                                appID: appId,
                                ...(noCache ? { noCache: true } : {}),
                            });
                            setNoCache(false);
                        }}
                    >
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            className="h-8 text-xs sm:text-sm"
                            isLoading={isDeploying}
                            disabled={isAppDeleting || (isAppActionPending && !isDeploying)}
                        >
                            <RefreshCw className="size-3.5 text-amber-500 dark:text-amber-400 mr-1 sm:mr-1.5" />
                            Re-deploy
                        </Button>
                    </PopConfirm>
                    <PopConfirm
                        title="Restart app"
                        description="Are you sure you want to restart this app?"
                        confirmText="Restart"
                        cancelText="Cancel"
                        variant="destructive"
                        onConfirm={() => {
                            restart({ projectID: projectId, env, appID: appId });
                        }}
                    >
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            className="h-8 text-xs sm:text-sm"
                            isLoading={isRestarting}
                            disabled={isAppDeleting || (isAppActionPending && !isRestarting)}
                        >
                            <Power className="size-3.5 text-amber-500 dark:text-amber-400 mr-1 sm:mr-1.5" />
                            Restart
                        </Button>
                    </PopConfirm>
                    {isAppStopped ? (
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            className="h-8 text-xs sm:text-sm"
                            isLoading={isSettingRunning}
                            disabled={isAppDeleting || (isAppActionPending && !isSettingRunning)}
                            onClick={() => {
                                setRunning({ projectID: projectId, env, appID: appId, running: true });
                            }}
                        >
                            <Power className="size-3.5 text-amber-500 dark:text-amber-400 mr-1 sm:mr-1.5" />
                            {startStopText}
                        </Button>
                    ) : (
                        <PopConfirm
                            title="Stop app"
                            description="Are you sure you want to stop this app?"
                            confirmText="Stop"
                            cancelText="Cancel"
                            variant="destructive"
                            onConfirm={() => {
                                setRunning({ projectID: projectId, env, appID: appId, running: false });
                            }}
                        >
                            <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                className="h-8 text-xs sm:text-sm"
                                isLoading={isSettingRunning}
                                disabled={isAppDeleting || (isAppActionPending && !isSettingRunning)}
                            >
                                <Power className="size-3.5 text-amber-500 dark:text-amber-400 mr-1 sm:mr-1.5" />
                                {startStopText}
                            </Button>
                        </PopConfirm>
                    )}
                </div>
            </div>

            {isFunctionApp(appData) && appData.runtimeOutdated && redeployedAppId !== appId ? (
                <div className="mb-2.5 flex flex-col sm:flex-row sm:items-center justify-between gap-2 rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-400">
                    <div>
                        <span className="font-medium">This function runs on an older runtime.</span> HivePaaS has been
                        updated since it was deployed: re-deploy it to build it on the current runtime.
                    </div>
                    <PopConfirm
                        title="Re-deploy function"
                        description="Build this function on the current runtime and deploy it?"
                        confirmText="Re-deploy"
                        cancelText="Cancel"
                        onConfirm={() => {
                            deploy({ projectID: projectId, env, appID: appId });
                        }}
                    >
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            className="h-7 shrink-0 text-xs"
                            isLoading={isDeploying}
                            disabled={isAppDeleting || (isAppActionPending && !isDeploying)}
                        >
                            <RefreshCw className="size-3.5 mr-1" />
                            Re-deploy
                        </Button>
                    </PopConfirm>
                </div>
            ) : null}

            <Separator className="opacity-50" />

            <div className="flex items-center gap-2 overflow-x-auto">
                <TabNavigation links={links} />
                <AppAccessLinksDropdown accessLinks={appData.accessLinks} />
            </div>
        </div>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
}

export const SingleAppHeader = memo(View);
