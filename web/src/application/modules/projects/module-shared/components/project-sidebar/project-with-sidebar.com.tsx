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
    label: string | React.ReactNode;
    disabled?: boolean;
}

function normalizePath(path: string) {
    return path.replace(/\/+$/, "");
}

function isRouteActive(route: string, pathname: string) {
    const normalizedRoute = normalizePath(route);
    const normalizedPathname = normalizePath(pathname);

    return normalizedPathname === normalizedRoute || normalizedPathname.startsWith(`${normalizedRoute}/`);
}

function createConfigurationTabs(projectId: string): TabItem[] {
    return [
        {
            label: "General",
            route: ROUTE.projects.single.configuration.general.$route(projectId),
        },
        {
            label: "Build Settings",
            route: ROUTE.projects.single.configuration.buildSettings.$route(projectId),
        },
        {
            label: "Storage Settings",
            route: ROUTE.projects.single.configuration.storageSettings.$route(projectId),
        },
        {
            label: "Domain Settings",
            route: ROUTE.projects.single.configuration.domainSettings.$route(projectId),
        },
        {
            label: "Danger Zone",
            route: ROUTE.projects.single.configuration.dangerZone.$route(projectId),
        },
    ];
}

function createProviderConfigurationTabs(projectId: string): TabItem[] {
    return [
        {
            label: "Access Tokens",
            route: ROUTE.projects.single.providerConfiguration.accessTokens.$route(projectId),
        },
        {
            label: "Basic Auth",
            route: ROUTE.projects.single.providerConfiguration.basicAuth.$route(projectId),
        },
        {
            label: "Cloud Storages",
            route: ROUTE.projects.single.providerConfiguration.cloudStorages.$route(projectId),
        },
        {
            label: "Email Accounts",
            route: ROUTE.projects.single.providerConfiguration.emailAccounts.$route(projectId),
        },
        {
            label: "Env Variables",
            route: ROUTE.projects.single.providerConfiguration.envVariables.$route(projectId),
        },
        {
            label: "Github Apps",
            route: ROUTE.projects.single.providerConfiguration.githubApps.$route(projectId),
        },
        {
            label: "IM Platforms",
            route: ROUTE.projects.single.providerConfiguration.imPlatforms.$route(projectId),
        },
        {
            label: "Notification Targets",
            route: ROUTE.projects.single.providerConfiguration.notificationTargets.$route(projectId),
        },
        {
            label: "Registry Auth",
            route: ROUTE.projects.single.providerConfiguration.registryAuth.$route(projectId),
        },
        {
            label: "Secrets",
            route: ROUTE.projects.single.providerConfiguration.secrets.$route(projectId),
        },
        {
            label: "SSH Keys",
            route: ROUTE.projects.single.providerConfiguration.sshKeys.$route(projectId),
        },
        {
            label: "SSL Certificates",
            route: ROUTE.projects.single.providerConfiguration.sslCertificates.$route(projectId),
        },
        {
            label: "Webhooks",
            route: ROUTE.projects.single.providerConfiguration.webhooks.$route(projectId),
        },
    ];
}

function createClusterResourcesTabs(projectId: string): TabItem[] {
    return [
        {
            label: "Networks",
            route: ROUTE.projects.single.clusterResources.networks.$route(projectId),
        },
    ];
}

function View({ projectId: projectIdProp, section = "providerConfiguration", children }: Props) {
    const { id: routeProjectId } = useParams<{ id: string }>();
    const location = useLocation();
    const projectId = projectIdProp ?? routeProjectId;

    invariant(projectId, "Project id must be defined");

    const tabs =
        section === "configuration"
            ? createConfigurationTabs(projectId)
            : section === "clusterResources"
              ? createClusterResourcesTabs(projectId)
              : createProviderConfigurationTabs(projectId);

    const activeKey = tabs.find(({ route }) => isRouteActive(route, location.pathname))?.route;
    return (
        <div className="flex flex-col gap-5 md:flex-row w-fit mx-auto">
            <aside className="md:w-56 md:shrink-0">
                <div className="sticky top-4">
                    <div className="bg-background rounded-lg py-4">
                        <Tabs
                            value={activeKey}
                            className="flex-row"
                        >
                            <TabsList className="bg-background h-full flex-col w-full rounded-none p-0">
                                {tabs.map(tab => (
                                    <TabsTrigger
                                        key={tab.route}
                                        value={tab.route}
                                        asChild
                                        className="py-3 pl-4 cursor-pointer bg-background data-[state=active]:border-primary data-[state=active]:bg-primary/10 dark:data-[state=active]:border-primary h-full w-full justify-start rounded-none border-0 border-l-2 border-transparent data-[state=active]:shadow-none"
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

interface Props extends PropsWithChildren {
    projectId?: string;
    section?: "configuration" | "providerConfiguration" | "clusterResources";
}

export const ProjectWithSidebar = memo(View);
