import { type PropsWithChildren, memo } from "react";

import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useLocation, useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface TabItem {
    route: string;
    label: string;
    disabled?: boolean;
}

function View({ children }: PropsWithChildren) {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "Project id must be defined");
    invariant(appId, "App id must be defined");

    const location = useLocation();

    const tabs: TabItem[] = [
        {
            label: "General",
            route: ROUTE.projects.single.apps.single.configuration.general.$route(projectId, appId),
        },
        {
            label: "Deployment Settings",
            route: ROUTE.projects.single.apps.single.configuration.deploymentSettings.$route(projectId, appId),
        },
        {
            label: "Container Settings",
            route: ROUTE.projects.single.apps.single.configuration.containerSettings.$route(projectId, appId),
        },
        {
            label: "HTTP Settings",
            route: ROUTE.projects.single.apps.single.configuration.httpSettings.$route(projectId, appId),
        },
        {
            label: "Env Variables",
            route: ROUTE.projects.single.apps.single.configuration.envVariables.$route(projectId, appId),
        },
        {
            label: "Secrets",
            route: ROUTE.projects.single.apps.single.configuration.secrets.$route(projectId, appId),
        },
        {
            label: "Availability & Scaling",
            route: ROUTE.projects.single.apps.single.configuration.availabilityAndScaling.$route(projectId, appId),
        },
        {
            label: "Persistent Storage",
            route: ROUTE.projects.single.apps.single.configuration.presistentStorage.$route(projectId, appId),
        },
        {
            label: "Networks",
            route: ROUTE.projects.single.apps.single.configuration.networks.$route(projectId, appId),
        },
        {
            label: "Resources",
            route: ROUTE.projects.single.apps.single.configuration.resources.$route(projectId, appId),
        },
        {
            label: "Danger Zone",
            route: ROUTE.projects.single.apps.single.configuration.dangerZone.$route(projectId, appId),
        },
    ];

    const activeKey = tabs.find(({ route }) => route === location.pathname)?.route;

    return (
        <div className="flex flex-col gap-5 md:flex-row w-fit mx-auto">
            <aside className="md:w-56 md:shrink-0">
                <div className="sticky top-4">
                    <div className="bg-background rounded-lg py-4">
                        <Tabs
                            defaultValue={activeKey}
                            className="flex-row"
                        >
                            <TabsList className="bg-background h-full flex-col w-full rounded-none p-0">
                                {tabs.map(tab => (
                                    <TabsTrigger
                                        key={tab.route}
                                        value={tab.route}
                                        asChild
                                        disabled={tab.disabled}
                                        className="py-3 cursor-pointer bg-background data-[state=active]:border-primary data-[state=active]:bg-primary/10 dark:data-[state=active]:border-primary h-full w-full justify-start rounded-none border-0 border-l-2 border-transparent data-[state=active]:shadow-none"
                                    >
                                        <AppLink.Basic
                                            className="w-full text-left"
                                            to={tab.route}
                                            ignorePrevPath
                                        >
                                            {tab.label}
                                        </AppLink.Basic>
                                    </TabsTrigger>
                                ))}
                            </TabsList>
                        </Tabs>
                    </div>
                </div>
            </aside>
            <div className={cn(listBox, "w-7xl")}>{children}</div>
        </div>
    );
}

export const SingleAppConfigurationLayout = memo(View);
