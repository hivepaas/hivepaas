import { memo } from "react";

import { Button } from "@components/ui";
import { Power, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectAppsCommands, ProjectAppsQueries, ProjectsQueries } from "~/projects/data";
import { ProjectAppStatusBadge, ProjectEnvBadge } from "~/projects/module-shared/components";

import { BackButton, TabNavigation } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { SingleAppBreadcrumbs } from "../buidling-blocks";

import { AppAccessLinksDropdown } from "./building-blocks";
import { SingleAppHeaderSkeleton } from "./single-app-header.skeleton.com";

function View({ projectId, appId }: Props) {
    const { data, isLoading, error } = ProjectsQueries.useFindOneById({ projectID: projectId });
    const {
        data: app,
        isLoading: isLoadingApp,
        error: errorApp,
    } = ProjectAppsQueries.useFindOneById({ projectID: projectId, appID: appId });
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
        ROUTE.projects.single.apps.single.configuration.dangerZone.$route(projectId, appId),
    ];

    const links = [
        {
            route: ROUTE.projects.single.apps.single.configuration.general.$route(projectId, appId),
            label: "Configuration",
            activePathPrefixes: configurationActivePathPrefixes,
        },
        {
            route: ROUTE.projects.single.apps.single.deployments.$route(projectId, appId),
            label: "Deployments",
        },
        {
            route: ROUTE.projects.single.apps.single.logs.$route(projectId, appId),
            label: "Logs",
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
                    <Button
                        type="button"
                        variant="outline"
                        isLoading={isDeploying}
                        disabled={isRestarting}
                        onClick={() => {
                            deploy({ projectID: projectId, appID: appId });
                        }}
                    >
                        <RefreshCw className="size-4 text-orange-600" />
                        Re-deploy
                    </Button>
                    <Button
                        type="button"
                        variant="outline"
                        isLoading={isRestarting}
                        disabled={isDeploying}
                        onClick={() => {
                            restart({ projectID: projectId, appID: appId });
                        }}
                    >
                        <Power className="size-4 text-orange-600" />
                        Restart
                    </Button>
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
