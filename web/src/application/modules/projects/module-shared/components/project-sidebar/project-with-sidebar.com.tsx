import { type PropsWithChildren, memo } from "react";

import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useLocation, useNavigate, useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface TabItem {
    route: string;
    label: string | React.ReactNode;
    disabled?: boolean;
}

function normalizePath(path: string) {
    return path.replace(/\/+$/, "");
}

function findActiveTab(tabs: TabItem[], pathname: string) {
    const normalizedPathname = normalizePath(pathname);
    const exactMatch = tabs.find(({ route }) => normalizePath(route) === normalizedPathname);
    if (exactMatch) {
        return exactMatch.route;
    }
    return tabs.find(({ route }) => normalizedPathname.startsWith(`${normalizePath(route)}/`))?.route;
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
            label: "ACME DNS Providers",
            route: ROUTE.projects.single.providerConfiguration.acmeDnsProviders.$route(projectId),
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
            label: "Command Pipes",
            route: ROUTE.projects.single.providerConfiguration.commandPipes.$route(projectId),
        },
        {
            label: "Command Templates",
            route: ROUTE.projects.single.providerConfiguration.commandTemplates.$route(projectId),
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
            label: "SSL Providers",
            route: ROUTE.projects.single.providerConfiguration.sslProviders.$route(projectId),
        },
    ];
}

function createSourcesTabs(projectId: string): TabItem[] {
    return [
        {
            label: "Github Apps",
            route: ROUTE.projects.single.sources.githubApps.$route(projectId),
        },
        {
            label: "Webhooks",
            route: ROUTE.projects.single.sources.webhooks.$route(projectId),
        },
    ];
}

function createClusterResourcesTabs(projectId: string): TabItem[] {
    return [
        {
            label: "Networks",
            route: ROUTE.projects.single.clusterResources.networks.$route(projectId),
        },
        {
            label: "Volumes",
            route: ROUTE.projects.single.clusterResources.volumes.$route(projectId),
        },
    ];
}

function View({ projectId: projectIdProp, section = "providerConfiguration", children }: Props) {
    const { id: routeProjectId } = useParams<{ id: string }>();
    const location = useLocation();
    const navigate = useNavigate();
    const projectId = projectIdProp ?? routeProjectId;

    invariant(projectId, "Project id must be defined");

    const tabs =
        section === "configuration"
            ? createConfigurationTabs(projectId)
            : section === "clusterResources"
              ? createClusterResourcesTabs(projectId)
              : section === "sources"
                ? createSourcesTabs(projectId)
                : createProviderConfigurationTabs(projectId);

    const activeKey = findActiveTab(tabs, location.pathname);
    return (
        <div className="flex flex-col gap-2 md:flex-row md:gap-5 w-full max-w-[1400px] mx-auto min-w-0">
            {/* Mobile Dropdown Navigation (< md) */}
            <div className="md:hidden w-full bg-background/95 backdrop-blur-md rounded-xl p-2 shadow-xs sticky top-[53px] z-20 border border-amber-500/25">
                <Select
                    value={activeKey}
                    onValueChange={val => {
                        if (val) {
                            void navigate(val);
                        }
                    }}
                >
                    <SelectTrigger className="w-full bg-gradient-to-r from-amber-500/10 via-amber-400/15 to-yellow-500/10 hover:from-amber-500/20 hover:to-yellow-500/20 border-amber-500/40 dark:border-amber-400/40 text-foreground font-semibold shadow-xs h-10 px-3.5 rounded-lg focus:ring-amber-400/40">
                        <SelectValue placeholder="Select section..." />
                    </SelectTrigger>
                    <SelectContent>
                        {tabs.map(tab => (
                            <SelectItem
                                key={tab.route}
                                value={tab.route}
                                disabled={tab.disabled}
                            >
                                {tab.label}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>

            {/* Desktop Vertical Sidebar Navigation (>= md) */}
            <aside className="hidden md:block w-56 shrink-0">
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
                                        disabled={tab.disabled}
                                        className="py-3 pl-4 cursor-pointer bg-background data-[state=active]:border-amber-500 dark:data-[state=active]:border-amber-400 data-[state=active]:bg-amber-500/[0.08] dark:data-[state=active]:bg-amber-400/[0.1] data-[state=active]:text-foreground font-medium h-full w-full justify-start rounded-none border-0 border-l-2 border-transparent data-[state=active]:shadow-none"
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
            <div className={cn(listBox, "w-full min-w-0 flex-1")}>{children}</div>
        </div>
    );
}

interface Props extends PropsWithChildren {
    projectId?: string;
    section?: "configuration" | "providerConfiguration" | "sources" | "clusterResources";
}

export const ProjectWithSidebar = memo(View);
