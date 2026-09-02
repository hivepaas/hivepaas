import { type PropsWithChildren, memo, useMemo } from "react";

import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import {
    AlertTriangle,
    Archive,
    Bell,
    Container,
    FileBadge,
    FileCode2,
    Fingerprint,
    Github,
    Globe,
    Hammer,
    HardDrive,
    KeyRound,
    Layers,
    Lock,
    type LucideIcon,
    Mail,
    MessageSquare,
    Network,
    Settings,
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

interface ProviderTabItem {
    route: string;
    label: string;
    icon: LucideIcon;
    isDanger?: boolean;
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

function createConfigurationSections(projectId: string): ProviderTabSection[] {
    return [
        {
            title: "Configuration",
            items: [
                {
                    label: "General",
                    icon: Settings,
                    route: ROUTE.projects.single.configuration.general.$route(projectId),
                },
                {
                    label: "Build Settings",
                    icon: Hammer,
                    route: ROUTE.projects.single.configuration.buildSettings.$route(projectId),
                },
                {
                    label: "Storage Settings",
                    icon: HardDrive,
                    route: ROUTE.projects.single.configuration.storageSettings.$route(projectId),
                },
                {
                    label: "Domain Settings",
                    icon: Globe,
                    route: ROUTE.projects.single.configuration.domainSettings.$route(projectId),
                },
            ],
        },
        {
            title: "Danger Zone",
            items: [
                {
                    label: "Danger Zone",
                    icon: AlertTriangle,
                    isDanger: true,
                    route: ROUTE.projects.single.configuration.dangerZone.$route(projectId),
                },
            ],
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
                    label: "Command Templates",
                    icon: FileCode2,
                    route: ROUTE.projects.single.providerConfiguration.commandTemplates.$route(projectId),
                },
                {
                    label: "Command Pipes",
                    icon: Workflow,
                    route: ROUTE.projects.single.providerConfiguration.commandPipes.$route(projectId),
                },
                {
                    label: "Backup Repos",
                    icon: Archive,
                    route: ROUTE.projects.single.providerConfiguration.backupRepos.$route(projectId),
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

function createClusterResourcesSections(projectId: string): ProviderTabSection[] {
    return [
        {
            title: "Cluster Resources",
            items: [
                {
                    label: "Networks",
                    icon: Network,
                    route: ROUTE.projects.single.clusterResources.networks.$route(projectId),
                },
                {
                    label: "Volumes",
                    icon: HardDrive,
                    route: ROUTE.projects.single.clusterResources.volumes.$route(projectId),
                },
            ],
        },
    ];
}

function View({ projectId: projectIdProp, section = "providerConfiguration", children }: Props) {
    const { id: routeProjectId } = useParams<{ id: string }>();
    const location = useLocation();
    const navigate = useNavigate();
    const projectId = projectIdProp ?? routeProjectId;

    invariant(projectId, "Project id must be defined");

    const sections = useMemo(() => {
        switch (section) {
            case "providerConfiguration":
                return createProviderConfigurationSections(projectId);
            case "clusterResources":
                return createClusterResourcesSections(projectId);
            case "configuration":
            default:
                return createConfigurationSections(projectId);
        }
    }, [section, projectId]);

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

interface Props extends PropsWithChildren {
    projectId?: string;
    section?: "configuration" | "providerConfiguration" | "clusterResources";
}

export const ProjectWithSidebar = memo(View);
