import { type PropsWithChildren, memo } from "react";

import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import {
    Bell,
    Container,
    FileBadge,
    FileCode2,
    Fingerprint,
    Github,
    Globe,
    HardDrive,
    KeyRound,
    Layers,
    Lock,
    type LucideIcon,
    Mail,
    MessageSquare,
    ShieldCheck,
    Terminal,
    Webhook,
    Workflow,
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
    label: string | React.ReactNode;
    disabled?: boolean;
}

interface ProviderTabItem {
    route: string;
    label: string;
    icon: LucideIcon;
}

interface ProviderTabSection {
    title: string;
    items: ProviderTabItem[];
}

function normalizePath(path: string) {
    return path.replace(/\/+$/, "");
}

function findActiveTab(tabs: { route: string }[], pathname: string) {
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

function createProviderConfigurationSections(projectId: string): ProviderTabSection[] {
    return [
        {
            title: "Sources",
            items: [
                {
                    label: "Github Apps",
                    // eslint-disable-next-line @typescript-eslint/no-deprecated -- lucide's Github brand icon is deprecated, kept intentionally here.
                    icon: Github,
                    route: ROUTE.projects.single.providerConfiguration.githubApps.$route(projectId),
                },
                {
                    label: "Webhooks",
                    icon: Webhook,
                    route: ROUTE.projects.single.providerConfiguration.webhooks.$route(projectId),
                },
            ],
        },
        {
            title: "Access & Authentication",
            items: [
                {
                    label: "Access Tokens",
                    icon: Fingerprint,
                    route: ROUTE.projects.single.providerConfiguration.accessTokens.$route(projectId),
                },
                {
                    label: "Basic Auth",
                    icon: Lock,
                    route: ROUTE.projects.single.providerConfiguration.basicAuth.$route(projectId),
                },
                {
                    label: "SSH Keys",
                    icon: Terminal,
                    route: ROUTE.projects.single.providerConfiguration.sshKeys.$route(projectId),
                },
                {
                    label: "Registry Auth",
                    icon: Container,
                    route: ROUTE.projects.single.providerConfiguration.registryAuth.$route(projectId),
                },
            ],
        },
        {
            title: "Domain & TLS",
            items: [
                {
                    label: "ACME DNS Providers",
                    icon: Globe,
                    route: ROUTE.projects.single.providerConfiguration.acmeDnsProviders.$route(projectId),
                },
                {
                    label: "SSL Certificates",
                    icon: FileBadge,
                    route: ROUTE.projects.single.providerConfiguration.sslCertificates.$route(projectId),
                },
                {
                    label: "SSL Providers",
                    icon: ShieldCheck,
                    route: ROUTE.projects.single.providerConfiguration.sslProviders.$route(projectId),
                },
            ],
        },
        {
            title: "Storages",
            items: [
                {
                    label: "Cloud Storages",
                    icon: HardDrive,
                    route: ROUTE.projects.single.providerConfiguration.cloudStorages.$route(projectId),
                },
            ],
        },
        {
            title: "Configuration",
            items: [
                {
                    label: "Env Variables",
                    icon: Layers,
                    route: ROUTE.projects.single.providerConfiguration.envVariables.$route(projectId),
                },
                {
                    label: "Secrets",
                    icon: KeyRound,
                    route: ROUTE.projects.single.providerConfiguration.secrets.$route(projectId),
                },
            ],
        },
        {
            title: "Automation",
            items: [
                {
                    label: "Command Pipes",
                    icon: Workflow,
                    route: ROUTE.projects.single.providerConfiguration.commandPipes.$route(projectId),
                },
                {
                    label: "Command Templates",
                    icon: FileCode2,
                    route: ROUTE.projects.single.providerConfiguration.commandTemplates.$route(projectId),
                },
            ],
        },
        {
            title: "Notifications",
            items: [
                {
                    label: "Email Accounts",
                    icon: Mail,
                    route: ROUTE.projects.single.providerConfiguration.emailAccounts.$route(projectId),
                },
                {
                    label: "IM Platforms",
                    icon: MessageSquare,
                    route: ROUTE.projects.single.providerConfiguration.imPlatforms.$route(projectId),
                },
                {
                    label: "Notification Targets",
                    icon: Bell,
                    route: ROUTE.projects.single.providerConfiguration.notificationTargets.$route(projectId),
                },
            ],
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

    if (section === "providerConfiguration") {
        const sections = createProviderConfigurationSections(projectId);
        const allTabs = sections.flatMap(sec => sec.items);
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
                            {sections.map((sec, sIndex) => (
                                <SelectGroup key={sec.title}>
                                    <SelectLabel className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground/70 px-2 py-1.5">
                                        {sec.title}
                                    </SelectLabel>
                                    {sec.items.map(tab => {
                                        const Icon = tab.icon;
                                        return (
                                            <SelectItem
                                                key={tab.route}
                                                value={tab.route}
                                                className="cursor-pointer"
                                            >
                                                <div className="flex items-center gap-2">
                                                    <Icon className="size-4 shrink-0 text-muted-foreground" />
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
                <aside className="hidden md:block w-56 shrink-0">
                    <div className="sticky top-4">
                        <div className="bg-background rounded-lg py-3 shadow-xs">
                            <Tabs
                                value={activeKey}
                                className="flex-row"
                            >
                                <TabsList className="bg-background h-full flex-col w-full rounded-none p-0 gap-3">
                                    {sections.map((sec, sIndex) => (
                                        <div
                                            key={sec.title}
                                            className="w-full flex flex-col"
                                        >
                                            <div className="px-3.5 pb-1.5 text-[11px] font-semibold tracking-wider text-muted-foreground/70 uppercase select-none">
                                                {sec.title}
                                            </div>
                                            <div className="flex flex-col">
                                                {sec.items.map(tab => {
                                                    const Icon = tab.icon;
                                                    return (
                                                        <TabsTrigger
                                                            key={tab.route}
                                                            value={tab.route}
                                                            asChild
                                                            className="py-2 px-3.5 cursor-pointer bg-background font-medium h-auto w-full justify-start rounded-none border-0 border-l-2 border-transparent transition-colors text-foreground hover:bg-muted/40 data-[state=active]:border-amber-500 dark:data-[state=active]:border-amber-400 data-[state=active]:bg-amber-500/[0.08] dark:data-[state=active]:bg-amber-400/[0.1] data-[state=active]:shadow-none [&[data-state=active]_.tab-icon]:text-amber-500 dark:[&[data-state=active]_.tab-icon]:text-amber-400"
                                                        >
                                                            <AppLink.Basic
                                                                className="w-full flex items-center gap-2.5 text-left text-sm"
                                                                to={tab.route}
                                                                ignorePrevPath
                                                            >
                                                                <Icon className="tab-icon size-4 shrink-0 transition-colors text-muted-foreground" />
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

    const tabs =
        section === "configuration" ? createConfigurationTabs(projectId) : createClusterResourcesTabs(projectId);

    const activeKey = findActiveTab(tabs, location.pathname);
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
    section?: "configuration" | "providerConfiguration" | "clusterResources";
}

export const ProjectWithSidebar = memo(View);
