import { type PropsWithChildren, memo, useMemo } from "react";

import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import {
    AlertTriangle,
    Box,
    CalendarClock,
    Copy,
    Cpu,
    FileCode2,
    FolderTree,
    Globe,
    HardDrive,
    KeyRound,
    Layers,
    type LucideIcon,
    Network,
    Repeat,
    Rocket,
    Settings,
    Sliders,
    TrendingUp,
} from "lucide-react";
import { useLocation, useNavigate, useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import {
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectLabel,
    SelectSeparator,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface TabItem {
    route: string;
    label: string;
    icon: LucideIcon;
    disabled?: boolean;
    isDanger?: boolean;
}

interface TabSection {
    title: string;
    items: TabItem[];
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

    const sections: TabSection[] = useMemo(
        () => [
            {
                title: "Deploy & Network",
                items: [
                    {
                        label: "General",
                        icon: Settings,
                        route: ROUTE.projects.single.apps.single.configuration.general.$route(projectId, env, appId),
                    },
                    {
                        label: "Deployment Settings",
                        icon: Rocket,
                        route: ROUTE.projects.single.apps.single.configuration.deploymentSettings.$route(
                            projectId,
                            env,
                            appId,
                        ),
                    },
                    {
                        label: "Routing Settings",
                        icon: Globe,
                        route: ROUTE.projects.single.apps.single.configuration.routingSettings.$route(
                            projectId,
                            env,
                            appId,
                        ),
                    },
                    {
                        label: "Networks",
                        icon: Network,
                        route: ROUTE.projects.single.apps.single.configuration.networks.$route(projectId, env, appId),
                    },
                ],
            },
            {
                title: "Config & Data",
                items: [
                    {
                        label: "Env Variables",
                        icon: Layers,
                        route: ROUTE.projects.single.apps.single.configuration.envVariables.$route(
                            projectId,
                            env,
                            appId,
                        ),
                    },
                    {
                        label: "Secrets",
                        icon: KeyRound,
                        route: ROUTE.projects.single.apps.single.configuration.secrets.$route(projectId, env, appId),
                    },
                    {
                        label: "Config Files",
                        icon: FileCode2,
                        route: ROUTE.projects.single.apps.single.configuration.configFiles.$route(
                            projectId,
                            env,
                            appId,
                        ),
                    },
                    {
                        label: "Data Files",
                        icon: FolderTree,
                        route: ROUTE.projects.single.apps.single.configuration.dataFiles.$route(projectId, env, appId),
                    },
                ],
            },
            {
                title: "Runtime & Scale",
                items: [
                    {
                        label: "Container Settings",
                        icon: Box,
                        route: ROUTE.projects.single.apps.single.configuration.containerSettings.$route(
                            projectId,
                            env,
                            appId,
                        ),
                    },
                    {
                        label: "Availability & Scaling",
                        icon: TrendingUp,
                        route: ROUTE.projects.single.apps.single.configuration.availabilityAndScaling.$route(
                            projectId,
                            env,
                            appId,
                        ),
                    },
                    {
                        label: "Persistent Storage",
                        icon: HardDrive,
                        route: ROUTE.projects.single.apps.single.configuration.presistentStorage.$route(
                            projectId,
                            env,
                            appId,
                        ),
                    },
                    {
                        label: "Resources",
                        icon: Cpu,
                        route: ROUTE.projects.single.apps.single.configuration.resources.$route(projectId, env, appId),
                    },
                ],
            },
            {
                title: "Automation",
                items: [
                    {
                        label: "Periodic Jobs",
                        icon: Repeat,
                        route: ROUTE.projects.single.apps.single.configuration.periodicJobs.$route(
                            projectId,
                            env,
                            appId,
                        ),
                    },
                    {
                        label: "Scheduled Jobs",
                        icon: CalendarClock,
                        route: ROUTE.projects.single.apps.single.configuration.scheduledJobs.$route(
                            projectId,
                            env,
                            appId,
                        ),
                    },
                ],
            },
            {
                title: "Operations",
                items: [
                    {
                        label: "App Clone",
                        icon: Copy,
                        route: ROUTE.projects.single.apps.single.configuration.appClone.$route(projectId, env, appId),
                    },
                    {
                        label: "Feature Settings",
                        icon: Sliders,
                        route: ROUTE.projects.single.apps.single.configuration.featureSettings.$route(
                            projectId,
                            env,
                            appId,
                        ),
                    },
                    {
                        label: "Danger Zone",
                        icon: AlertTriangle,
                        isDanger: true,
                        route: ROUTE.projects.single.apps.single.configuration.dangerZone.$route(projectId, env, appId),
                    },
                ],
            },
        ],
        [projectId, env, appId],
    );

    const allTabs = useMemo(() => sections.flatMap(sec => sec.items), [sections]);
    const activeKey = findActiveTab(allTabs, location.pathname);

    return (
        <div className="flex flex-col gap-2 md:flex-row md:gap-4 w-full max-w-[1400px] mx-auto min-w-0">
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
                        {sections.map((section, sIndex) => (
                            <SelectGroup key={section.title}>
                                <SelectLabel className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground/70 px-2 py-1.5">
                                    {section.title}
                                </SelectLabel>
                                {section.items.map(tab => {
                                    const Icon = tab.icon;
                                    return (
                                        <SelectItem
                                            key={tab.route}
                                            value={tab.route}
                                            disabled={tab.disabled}
                                            className={cn(
                                                "cursor-pointer",
                                                tab.isDanger &&
                                                    "text-destructive focus:text-destructive focus:bg-destructive/10",
                                            )}
                                        >
                                            <div className="flex items-center gap-2">
                                                <Icon
                                                    className={cn(
                                                        "size-4 shrink-0",
                                                        tab.isDanger ? "text-destructive" : "text-muted-foreground",
                                                    )}
                                                />
                                                <span>{tab.label}</span>
                                            </div>
                                        </SelectItem>
                                    );
                                })}
                                {sIndex < sections.length - 1 && <SelectSeparator />}
                            </SelectGroup>
                        ))}
                    </SelectContent>
                </Select>
            </div>

            {/* Desktop Vertical Sidebar Navigation (>= md) */}
            <aside className="hidden md:block w-[236px] shrink-0">
                <div className="sticky top-4">
                    <div className="bg-background rounded-lg py-3 shadow-xs">
                        <Tabs
                            value={activeKey}
                            className="flex-row"
                        >
                            <TabsList className="bg-background h-full flex-col w-full rounded-none p-0 gap-3">
                                {sections.map((section, sIndex) => (
                                    <div
                                        key={section.title}
                                        className="w-full flex flex-col"
                                    >
                                        <div className="px-3.5 pb-1.5 text-[11px] font-semibold tracking-wider text-muted-foreground/70 uppercase select-none">
                                            {section.title}
                                        </div>
                                        <div className="flex flex-col">
                                            {section.items.map(tab => {
                                                const Icon = tab.icon;
                                                return (
                                                    <TabsTrigger
                                                        key={tab.route}
                                                        value={tab.route}
                                                        asChild
                                                        disabled={tab.disabled}
                                                        className={cn(
                                                            "py-2 px-3.5 cursor-pointer bg-background font-medium h-auto w-full justify-start rounded-none border-0 border-l-2 border-transparent transition-colors",
                                                            tab.isDanger
                                                                ? "text-destructive hover:bg-destructive/10 hover:text-destructive data-[state=active]:border-destructive data-[state=active]:bg-destructive/10 data-[state=active]:text-destructive data-[state=active]:shadow-none"
                                                                : "text-foreground hover:bg-muted/40 data-[state=active]:border-amber-500 dark:data-[state=active]:border-amber-400 data-[state=active]:bg-amber-500/[0.08] dark:data-[state=active]:bg-amber-400/[0.1] data-[state=active]:shadow-none [&[data-state=active]_.tab-icon]:text-amber-500 dark:[&[data-state=active]_.tab-icon]:text-amber-400",
                                                        )}
                                                    >
                                                        <AppLink.Basic
                                                            className="w-full flex items-center gap-2.5 text-left text-sm"
                                                            to={tab.route}
                                                            ignorePrevPath
                                                        >
                                                            <Icon
                                                                className={cn(
                                                                    "tab-icon size-4 shrink-0 transition-colors",
                                                                    tab.isDanger
                                                                        ? "text-destructive"
                                                                        : "text-muted-foreground",
                                                                )}
                                                            />
                                                            <span className="truncate">{tab.label}</span>
                                                        </AppLink.Basic>
                                                    </TabsTrigger>
                                                );
                                            })}
                                        </div>
                                        {sIndex < sections.length - 1 && (
                                            <div className="pt-2 px-3.5">
                                                <Separator className="opacity-40" />
                                            </div>
                                        )}
                                    </div>
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
