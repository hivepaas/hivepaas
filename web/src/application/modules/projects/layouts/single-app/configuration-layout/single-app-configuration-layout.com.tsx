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
    label: string;
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

function View({ children }: PropsWithChildren) {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();

    invariant(projectId, "Project id must be defined");
    invariant(env, "Environment must be defined");
    invariant(appId, "App id must be defined");

    const location = useLocation();
    const navigate = useNavigate();

    const tabs: TabItem[] = [
        {
            label: "General",
            route: ROUTE.projects.single.apps.single.configuration.general.$route(projectId, env, appId),
        },
        {
            label: "Deployment Settings",
            route: ROUTE.projects.single.apps.single.configuration.deploymentSettings.$route(projectId, env, appId),
        },
        {
            label: "Routing Settings",
            route: ROUTE.projects.single.apps.single.configuration.routingSettings.$route(projectId, env, appId),
        },

        {
            label: "Env Variables",
            route: ROUTE.projects.single.apps.single.configuration.envVariables.$route(projectId, env, appId),
        },
        {
            label: "Secrets",
            route: ROUTE.projects.single.apps.single.configuration.secrets.$route(projectId, env, appId),
        },
        {
            label: "Config Files",
            route: ROUTE.projects.single.apps.single.configuration.configFiles.$route(projectId, env, appId),
        },
        {
            label: "Data Files",
            route: ROUTE.projects.single.apps.single.configuration.dataFiles.$route(projectId, env, appId),
        },
        {
            label: "Container Settings",
            route: ROUTE.projects.single.apps.single.configuration.containerSettings.$route(projectId, env, appId),
        },
        {
            label: "Availability & Scaling",
            route: ROUTE.projects.single.apps.single.configuration.availabilityAndScaling.$route(projectId, env, appId),
        },
        {
            label: "Persistent Storage",
            route: ROUTE.projects.single.apps.single.configuration.presistentStorage.$route(projectId, env, appId),
        },
        {
            label: "Networks",
            route: ROUTE.projects.single.apps.single.configuration.networks.$route(projectId, env, appId),
        },
        {
            label: "Resources",
            route: ROUTE.projects.single.apps.single.configuration.resources.$route(projectId, env, appId),
        },
        {
            label: "Periodic Jobs",
            route: ROUTE.projects.single.apps.single.configuration.periodicJobs.$route(projectId, env, appId),
        },
        {
            label: "Scheduled Jobs",
            route: ROUTE.projects.single.apps.single.configuration.scheduledJobs.$route(projectId, env, appId),
        },
        {
            label: "App Clone",
            route: ROUTE.projects.single.apps.single.configuration.appClone.$route(projectId, env, appId),
        },
        {
            label: "Feature Settings",
            route: ROUTE.projects.single.apps.single.configuration.featureSettings.$route(projectId, env, appId),
        },
        {
            label: "Danger Zone",
            route: ROUTE.projects.single.apps.single.configuration.dangerZone.$route(projectId, env, appId),
        },
    ];

    const activeKey = findActiveTab(tabs, location.pathname);

    return (
        <div className="flex flex-col gap-2 md:flex-row md:gap-5 w-full max-w-[1400px] mx-auto min-w-0">
            {/* Mobile Dropdown Navigation (< md) */}
            <div className="md:hidden w-full bg-background rounded-lg p-2.5 shadow-xs">
                <Select
                    value={activeKey}
                    onValueChange={val => {
                        if (val) {
                            void navigate(val);
                        }
                    }}
                >
                    <SelectTrigger className="w-full bg-muted/40 font-medium">
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
            <div className={cn(listBox, "w-full min-w-0 flex-1")}>{children}</div>
        </div>
    );
}

export const SingleAppConfigurationLayout = memo(View);
