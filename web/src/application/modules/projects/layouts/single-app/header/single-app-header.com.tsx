import { memo } from "react";

import { Button } from "@components/ui";
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

function View({ projectId, appId }: Props) {
    const { scheduledJobId, taskId } = useParams<{
        scheduledJobId?: string;
        taskId?: string;
    }>();
    const { data, isLoading, error } = ProjectsQueries.useFindOneById({ projectID: projectId });
    const {
        data: app,
        isLoading: isLoadingApp,
        error: errorApp,
    } = ProjectAppsQueries.useFindOneById({ projectID: projectId, appID: appId });
    const { data: scheduledJobResponse } = AppScheduledJobsQueries.useFindOneById(
        {
            projectID: projectId,
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
    const appEnv = project.envs.find(env => env.name === appData.env);
    const appRoute = ROUTE.projects.single.apps.single.configuration.general.$route(projectId, appId);
    const scheduledJobName = scheduledJobResponse?.data.name.trim();
    const scheduledJobTasksLabel = scheduledJobName ? `${scheduledJobName} Tasks` : "Scheduled Job Tasks";
    const taskBreadcrumbItems = scheduledJobId
        ? [
              {
                  label: "Scheduled Jobs",
                  to: ROUTE.projects.single.apps.single.configuration.scheduledJobs.$route(projectId, appId),
              },
              {
                  label: scheduledJobTasksLabel,
                  ...(taskId
                      ? {
                            to: ROUTE.projects.single.apps.single.scheduledJobTasks.$route(
                                projectId,
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
        ROUTE.projects.single.apps.single.configuration.deploymentSettings.$route(projectId, appId),
        ROUTE.projects.single.apps.single.configuration.httpSettings.$route(projectId, appId),
        ROUTE.projects.single.apps.single.configuration.envVariables.$route(projectId, appId),
        ROUTE.projects.single.apps.single.configuration.secrets.$route(projectId, appId),
        ROUTE.projects.single.apps.single.configuration.configFiles.$route(projectId, appId),
        ROUTE.projects.single.apps.single.configuration.containerSettings.$route(projectId, appId),
        ROUTE.projects.single.apps.single.configuration.availabilityAndScaling.$route(projectId, appId),
        ROUTE.projects.single.apps.single.configuration.presistentStorage.$route(projectId, appId),
        ROUTE.projects.single.apps.single.configuration.networks.$route(projectId, appId),
        ROUTE.projects.single.apps.single.configuration.resources.$route(projectId, appId),
        ROUTE.projects.single.apps.single.configuration.healthChecks.$route(projectId, appId),
        ROUTE.projects.single.apps.single.configuration.scheduledJobs.$route(projectId, appId),
        ROUTE.projects.single.apps.single.configuration.featureSettings.$route(projectId, appId),
        ROUTE.projects.single.apps.single.configuration.dangerZone.$route(projectId, appId),
    ];

    const links = [
        {
            route: ROUTE.projects.single.apps.single.configuration.general.$route(projectId, appId),
            label: "Settings",
            activePathPrefixes: configurationActivePathPrefixes,
        },
        {
            route: ROUTE.projects.single.apps.single.instances.$route(projectId, appId),
            label: "Instances",
            activePathPrefixes: [ROUTE.projects.single.apps.single.instances.$route(projectId, appId)],
        },
        {
            route: ROUTE.projects.single.apps.single.deployments.$route(projectId, appId),
            label: "Deployments",
            activePathPrefixes: [ROUTE.projects.single.apps.single.deployments.$route(projectId, appId)],
        },
        {
            route: ROUTE.projects.single.apps.single.logs.$route(projectId, appId),
            label: "Logs",
        },
        {
            route: ROUTE.projects.single.apps.single.terminal.$route(projectId, appId),
            label: "Terminal",
        },
        {
            route: ROUTE.projects.single.apps.single.previewDeployments.$route(projectId, appId),
            label: "Preview Deployments",
        },
    ];
    return (
        <div className="bg-background pt-4 px-5 rounded-lg">
            <div className="flex items-center justify-between">
                <SingleAppBreadcrumbs
                    app={appData}
                    appRoute={appRoute}
                    items={taskBreadcrumbItems}
                    project={project}
                />
            </div>

            <div className="flex items-center gap-4 mt-4 pb-4">
                <BackButton />
                <div className="flex items-center gap-4">
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
                            deploy({ projectID: projectId, appID: appId });
                        }}
                    >
                        <Button
                            type="button"
                            variant="outline"
                            isLoading={isDeploying}
                            disabled={isRestarting}
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
                            restart({ projectID: projectId, appID: appId });
                        }}
                    >
                        <Button
                            type="button"
                            variant="outline"
                            isLoading={isRestarting}
                            disabled={isDeploying}
                        >
                            <Power className="size-4 text-orange-600" />
                            Restart
                        </Button>
                    </PopConfirm>
                </div>
            </div>
        </div>
    );
}

interface Props {
    projectId: string;
    appId: string;
}

export const SingleAppHeader = memo(View);
