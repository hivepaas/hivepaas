import { memo } from "react";

import { Button } from "@components/ui";
import { Avatar } from "@components/ui/avatar";
import { Power, RefreshCw } from "lucide-react";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppScheduledJobsQueries, ProjectAppsCommands, ProjectAppsQueries, ProjectsQueries } from "~/projects/data";
import { ProjectAppStatusBadge, ProjectEnvBadge } from "~/projects/module-shared/components";

import { BackButton, PopConfirm, TabNavigation } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { SingleAppBreadcrumbs } from "../buidling-blocks";

import { AppAccessLinksDropdown } from "./building-blocks";
import { SingleAppHeaderSkeleton } from "./single-app-header.skeleton.com";

function View({ projectId, env, appId }: Props) {
    const { scheduledJobId, taskId } = useParams<{
        scheduledJobId?: string;
        taskId?: string;
    }>();
    const { data, isLoading, error } = ProjectsQueries.useFindOneById({ projectID: projectId });
    const {
        data: app,
        isLoading: isLoadingApp,
        error: errorApp,
    } = ProjectAppsQueries.useFindOneById({ projectID: projectId, env, appID: appId, getStats: true });
    const { data: scheduledJobResponse } = AppScheduledJobsQueries.useFindOneById(
        {
            projectID: projectId,
            env,
            appID: appId,
            scheduledJobID: scheduledJobId ?? "",
        },
        {
            enabled: Boolean(scheduledJobId),
        },
    );
    const { mutate: deploy, isPending: isDeploying } = ProjectAppsCommands.useDeploy({
        onSuccess: () => {
            toast.success("Re-deploy started");
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
    const isAppRunning = (appData.stats?.runningTasks ?? 0) > 0;
    const startStopText = isAppRunning ? "Stop" : "Start";
    const appEnv = project.envs.find(projectEnv => projectEnv.name === appData.env);
    const appRoute = ROUTE.projects.single.apps.single.configuration.general.$route(projectId, env, appId);
    const scheduledJobName = scheduledJobResponse?.data.name.trim();
    const scheduledJobTasksLabel = scheduledJobName ? `${scheduledJobName} Tasks` : "Scheduled Job Tasks";
    const taskBreadcrumbItems = scheduledJobId
        ? [
              {
                  label: "Scheduled Jobs",
                  to: ROUTE.projects.single.apps.single.configuration.scheduledJobs.$route(projectId, env, appId),
              },
              {
                  label: scheduledJobTasksLabel,
                  ...(taskId
                      ? {
                            to: ROUTE.projects.single.apps.single.scheduledJobTasks.$route(
                                projectId,
                                env,
                                appId,
                                scheduledJobId,
                            ),
                        }
                      : {}),
              },
              ...(taskId
                  ? [
                        {
                            label: "Task Details",
                        },
                    ]
                  : []),
          ]
        : [];
    const configurationActivePathPrefixes = [
        ROUTE.projects.single.apps.single.configuration.deploymentSettings.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.httpSettings.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.envVariables.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.secrets.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.configFiles.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.containerSettings.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.availabilityAndScaling.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.presistentStorage.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.networks.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.resources.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.periodicJobs.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.scheduledJobs.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.featureSettings.$route(projectId, env, appId),
        ROUTE.projects.single.apps.single.configuration.dangerZone.$route(projectId, env, appId),
    ];

    const links = [
        {
            route: ROUTE.projects.single.apps.single.configuration.general.$route(projectId, env, appId),
            label: "Settings",
            activePathPrefixes: configurationActivePathPrefixes,
        },
        {
            route: ROUTE.projects.single.apps.single.instances.$route(projectId, env, appId),
            label: "Instances",
            activePathPrefixes: [ROUTE.projects.single.apps.single.instances.$route(projectId, env, appId)],
        },
        {
            route: ROUTE.projects.single.apps.single.deployments.$route(projectId, env, appId),
            label: "Deployments",
            activePathPrefixes: [ROUTE.projects.single.apps.single.deployments.$route(projectId, env, appId)],
        },
        {
            route: ROUTE.projects.single.apps.single.logs.$route(projectId, env, appId),
            label: "Logs",
        },
        {
            route: ROUTE.projects.single.apps.single.terminal.$route(projectId, env, appId),
            label: "Terminal",
        },
        ...(!isChildApp
            ? [
                  {
                      route: ROUTE.projects.single.apps.single.previewDeployments.$route(projectId, env, appId),
                      label: "Preview Deployments",
                  },
              ]
            : []),
    ];
    return (
        <div className="bg-background pt-4 px-5 rounded-lg">
            <div className="flex items-center justify-between">
                <SingleAppBreadcrumbs
                    app={appData}
                    appRoute={appRoute}
                    items={taskBreadcrumbItems}
                    parentApp={appData.parentApp}
                    project={project}
                />
            </div>

            <div className="flex items-center gap-4 mt-4 pb-4">
                <BackButton />
                <div className="flex items-center gap-4">
                    <Avatar
                        name={appData.name}
                        src={appData.photo}
                        className="size-10 text-base"
                    />
                    <div className="flex flex-col gap-3">
                        <div className="flex items-center gap-2">
                            <h2 className="text-[20px] font-semibold text-foreground">{appData.name}</h2>
                            <ProjectAppStatusBadge status={appData.status} />
                            {appData.env && (
                                <ProjectEnvBadge
                                    name={appData.env}
                                    color={appEnv?.color}
                                />
                            )}
                        </div>
                    </div>
                </div>
            </div>

            <div className="border-b border-border" />

            <div className="flex flex-wrap items-center gap-2">
                <div className="flex min-w-0 flex-1 items-center gap-2">
                    <TabNavigation links={links} />
                    <AppAccessLinksDropdown accessLinks={appData.accessLinks} />
                </div>

                <div className="ml-auto flex items-center gap-2 pb-1">
                    <PopConfirm
                        title="Re-deploy app"
                        description="Are you sure you want to re-deploy this app?"
                        confirmText="Re-deploy"
                        cancelText="Cancel"
                        onConfirm={() => {
                            deploy({ projectID: projectId, env, appID: appId });
                        }}
                    >
                        <Button
                            type="button"
                            variant="outline"
                            isLoading={isDeploying}
                            disabled={isAppActionPending && !isDeploying}
                        >
                            <RefreshCw className="size-4 text-orange-600" />
                            Re-deploy
                        </Button>
                    </PopConfirm>
                    <PopConfirm
                        title="Restart app"
                        description="Are you sure you want to restart this app?"
                        confirmText="Restart"
                        cancelText="Cancel"
                        onConfirm={() => {
                            restart({ projectID: projectId, env, appID: appId });
                        }}
                    >
                        <Button
                            type="button"
                            variant="outline"
                            isLoading={isRestarting}
                            disabled={isAppActionPending && !isRestarting}
                        >
                            <Power className="size-4 text-orange-600" />
                            Restart
                        </Button>
                    </PopConfirm>
                    {isAppRunning ? (
                        <PopConfirm
                            title="Stop app"
                            description="Are you sure you want to stop this app?"
                            confirmText="Stop"
                            cancelText="Cancel"
                            onConfirm={() => {
                                setRunning({ projectID: projectId, env, appID: appId, running: false });
                            }}
                        >
                            <Button
                                type="button"
                                variant="outline"
                                isLoading={isSettingRunning}
                                disabled={isAppActionPending && !isSettingRunning}
                            >
                                <Power className="size-4 text-orange-600" />
                                {startStopText}
                            </Button>
                        </PopConfirm>
                    ) : (
                        <Button
                            type="button"
                            variant="outline"
                            isLoading={isSettingRunning}
                            disabled={isAppActionPending && !isSettingRunning}
                            onClick={() => {
                                setRunning({ projectID: projectId, env, appID: appId, running: true });
                            }}
                        >
                            <Power className="size-4 text-orange-600" />
                            {startStopText}
                        </Button>
                    )}
                </div>
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
