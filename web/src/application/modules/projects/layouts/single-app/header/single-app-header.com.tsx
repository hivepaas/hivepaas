import { memo } from "react";

import invariant from "tiny-invariant";
import { ProjectAppsQueries, ProjectsQueries } from "~/projects/data";
import { ProjectAppStatusBadge } from "~/projects/module-shared/components";

import { BackButton, TabNavigation } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { SingleAppBreadcrumbs } from "../buidling-blocks";

import { SingleAppHeaderSkeleton } from "./single-app-header.skeleton.com";

function View({ projectId, appId }: Props) {
    const { data, isLoading, error } = ProjectsQueries.useFindOneById({ projectID: projectId });
    const {
        data: app,
        isLoading: isLoadingApp,
        error: errorApp,
    } = ProjectAppsQueries.useFindOneById({ projectID: projectId, appID: appId });

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

    const links = [
        {
            route: ROUTE.projects.single.apps.single.configuration.general.$route(projectId, appId),
            label: "Configuration",
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
                        </div>
                    </div>
                </div>
            </div>

            <div className="border-b border-border" />

            <TabNavigation links={links} />
        </div>
    );
}

interface Props {
    projectId: string;
    appId: string;
}

export const SingleAppHeader = memo(View);
