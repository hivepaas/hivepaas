import * as React from "react";

import { LogoIcon } from "@/assets/icons";
import {
    ArchiveX,
    Bell,
    CircleDashed,
    CircleUser,
    Container,
    DatabaseBackup,
    FileBadge,
    Fingerprint,
    Github,
    Globe,
    Hammer,
    HardDrive,
    Key,
    KeyRound,
    Layers,
    LayoutGrid,
    ListTodo,
    Lock,
    type LucideIcon,
    Mail,
    MapPin,
    MessageSquare,
    Network,
    Puzzle,
    RefreshCw,
    Route,
    Server,
    Settings,
    Settings2,
    ShieldCheck,
    Terminal,
    Trash2,
    User,
    Users,
    Webhook,
} from "lucide-react";
import { useLocation } from "react-router";

import { MODULE_IDS, ROUTE, type ResourceModuleId } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";
import { type ModuleId, type ModulePermission, useConditionalModuleCollections } from "@application/shared/permissions";

import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarHeader,
    SidebarRail,
    useSidebar,
} from "@/components/ui/sidebar";

import { NavMain } from "../nav-main";
import { NavUser } from "../nav-user";

interface SidebarSubItem {
    title: string;
    route: string;
    pattern: string;
    icon?: LucideIcon;
    moduleId?: ResourceModuleId;
}

interface SidebarSection {
    title: string;
    items: SidebarSubItem[];
}

interface SidebarItem {
    title: string;
    route: string;
    pattern: string;
    icon?: LucideIcon;
    moduleId?: ResourceModuleId;
    alwaysVisible?: boolean;
    items?: SidebarSubItem[];
    sections?: SidebarSection[];
}

const navMain: SidebarItem[] = [
    {
        title: "Projects",
        route: ROUTE.projects.list.$route,
        pattern: ROUTE.projects.list.$pattern,
        icon: LayoutGrid,
        moduleId: MODULE_IDS.Project,
        alwaysVisible: true,
    },
    {
        title: "Integrations",
        route: "#",
        pattern: ROUTE.settings.$pattern,
        icon: Puzzle,
        moduleId: MODULE_IDS.Settings,
        sections: [
            {
                title: "Sources",
                items: [
                    {
                        title: "Github Apps",
                        // eslint-disable-next-line @typescript-eslint/no-deprecated -- lucide's Github brand icon is deprecated, kept intentionally here.
                        icon: Github,
                        route: ROUTE.settings.githubApps.$route,
                        pattern: ROUTE.settings.githubApps.$pattern,
                    },
                    {
                        title: "Webhooks",
                        icon: Webhook,
                        route: ROUTE.settings.webhooks.$route,
                        pattern: ROUTE.settings.webhooks.$pattern,
                    },
                ],
            },
            {
                title: "Access & Authentication",
                items: [
                    {
                        title: "Access Tokens",
                        icon: Fingerprint,
                        route: ROUTE.settings.accessTokens.$route,
                        pattern: ROUTE.settings.accessTokens.$pattern,
                    },
                    {
                        title: "Basic Auth",
                        icon: Lock,
                        route: ROUTE.settings.basicAuth.$route,
                        pattern: ROUTE.settings.basicAuth.$pattern,
                    },
                    {
                        title: "SSH Keys",
                        icon: Terminal,
                        route: ROUTE.settings.sshKeys.$route,
                        pattern: ROUTE.settings.sshKeys.$pattern,
                    },
                    {
                        title: "Registry Auth",
                        icon: Container,
                        route: ROUTE.settings.registryAuth.$route,
                        pattern: ROUTE.settings.registryAuth.$pattern,
                    },
                    {
                        title: "OAuth",
                        icon: KeyRound,
                        route: ROUTE.settings.oauth.$route,
                        pattern: ROUTE.settings.oauth.$pattern,
                    },
                ],
            },
            {
                title: "Domain & TLS",
                items: [
                    {
                        title: "ACME DNS Providers",
                        icon: Globe,
                        route: ROUTE.settings.acmeDnsProviders.$route,
                        pattern: ROUTE.settings.acmeDnsProviders.$pattern,
                    },
                    {
                        title: "SSL Certificates",
                        icon: FileBadge,
                        route: ROUTE.settings.sslCertificates.$route,
                        pattern: ROUTE.settings.sslCertificates.$pattern,
                    },
                    {
                        title: "SSL Providers",
                        icon: ShieldCheck,
                        route: ROUTE.settings.sslProviders.$route,
                        pattern: ROUTE.settings.sslProviders.$pattern,
                    },
                ],
            },
            {
                title: "Storages",
                items: [
                    {
                        title: "Cloud Storages",
                        icon: HardDrive,
                        route: ROUTE.settings.cloudStorages.$route,
                        pattern: ROUTE.settings.cloudStorages.$pattern,
                    },
                ],
            },
            {
                title: "Notifications",
                items: [
                    {
                        title: "Email Accounts",
                        icon: Mail,
                        route: ROUTE.settings.emailAccounts.$route,
                        pattern: ROUTE.settings.emailAccounts.$pattern,
                    },
                    {
                        title: "IM Platforms",
                        icon: MessageSquare,
                        route: ROUTE.settings.imPlatforms.$route,
                        pattern: ROUTE.settings.imPlatforms.$pattern,
                    },
                    {
                        title: "Notification Targets",
                        icon: Bell,
                        route: ROUTE.settings.notificationTargets.$route,
                        pattern: ROUTE.settings.notificationTargets.$pattern,
                    },
                ],
            },
        ],
    },
    {
        title: "Settings",
        route: "#",
        pattern: ROUTE.appSettings.$pattern,
        icon: Settings,
        moduleId: MODULE_IDS.Settings,
        items: [
            {
                title: "Image Build",
                icon: Hammer,
                route: ROUTE.appSettings.imageBuild.$route,
                pattern: ROUTE.appSettings.imageBuild.$pattern,
            },
            {
                title: "App Placement",
                icon: MapPin,
                route: ROUTE.appSettings.appPlacement.$route,
                pattern: ROUTE.appSettings.appPlacement.$pattern,
            },
        ],
    },
    {
        title: "Cluster",
        route: "#",
        pattern: "#",
        icon: Container,
        moduleId: MODULE_IDS.Cluster,
        items: [
            {
                title: "Nodes",
                icon: Server,
                route: ROUTE.cluster.nodes.$route,
                pattern: ROUTE.cluster.nodes.$pattern,
            },
            {
                title: "Networks",
                icon: Network,
                route: ROUTE.cluster.networks.$route,
                pattern: ROUTE.cluster.networks.$pattern,
            },
            {
                title: "Volumes",
                icon: HardDrive,
                route: ROUTE.cluster.volumes.$route,
                pattern: ROUTE.cluster.volumes.$pattern,
            },
        ],
    },
    {
        title: "System",
        route: "#",
        pattern: ROUTE.systemSettings.$pattern,
        icon: Settings2,
        moduleId: MODULE_IDS.System,
        items: [
            {
                title: "HivePaaS",
                icon: Layers,
                route: ROUTE.systemSettings.hivepaas.general.$route,
                pattern: ROUTE.systemSettings.hivepaas.$pattern,
            },
            {
                title: "Traefik",
                icon: Route,
                route: ROUTE.systemSettings.traefik.general.$route,
                pattern: ROUTE.systemSettings.traefik.$pattern,
            },
            {
                title: "Backup Repo Cleanup",
                icon: ArchiveX,
                route: ROUTE.systemSettings.backupRepoCleanup.configuration.$route,
                pattern: ROUTE.systemSettings.backupRepoCleanup.$pattern,
            },
            {
                title: "SSL Renewal",
                icon: RefreshCw,
                route: ROUTE.systemSettings.sslRenewal.configuration.$route,
                pattern: ROUTE.systemSettings.sslRenewal.$pattern,
            },
            {
                title: "Data Backup",
                icon: DatabaseBackup,
                route: ROUTE.systemSettings.dataBackup.configuration.$route,
                pattern: ROUTE.systemSettings.dataBackup.$pattern,
            },
            {
                title: "Data Cleanup",
                icon: Trash2,
                route: ROUTE.systemSettings.dataCleanup.configuration.$route,
                pattern: ROUTE.systemSettings.dataCleanup.$pattern,
            },
        ],
    },
    {
        title: "System Status",
        route: "#",
        pattern: ROUTE.systemStatus.$pattern,
        icon: CircleDashed,
        moduleId: MODULE_IDS.System,
        items: [
            {
                title: "Tasks",
                icon: ListTodo,
                route: ROUTE.systemStatus.tasks.$route,
                pattern: ROUTE.systemStatus.tasks.$pattern,
            },
        ],
    },
    {
        title: "User Management",
        route: ROUTE.userManagement.users.$route,
        pattern: ROUTE.userManagement.users.$pattern,
        icon: Users,
        moduleId: MODULE_IDS.User,
    },
    {
        title: "Your Account",
        route: "#",
        pattern: "#",
        icon: User,
        items: [
            {
                title: "Account",
                icon: CircleUser,
                route: ROUTE.currentUser.profile.$route,
                pattern: ROUTE.currentUser.profile.$pattern,
            },
            {
                title: "API Keys",
                icon: Key,
                route: ROUTE.currentUser.profileApiKeys.$route,
                pattern: ROUTE.currentUser.profileApiKeys.$pattern,
            },
        ],
    },
];

function hasReadableModuleAccess(
    item: { moduleId?: ResourceModuleId; alwaysVisible?: boolean },
    permissions: ReadonlyMap<ModuleId, ModulePermission>,
) {
    return (item.alwaysVisible ?? false) || !item.moduleId || permissions.get(item.moduleId)?.actions.read === true;
}

function filterSidebarItems(
    items: readonly SidebarItem[],
    permissions: ReadonlyMap<ModuleId, ModulePermission>,
): SidebarItem[] {
    return items.flatMap((item): SidebarItem[] => {
        if (!hasReadableModuleAccess(item, permissions)) {
            return [];
        }

        if (item.sections && item.sections.length > 0) {
            const filteredSections = item.sections.flatMap(sec => {
                const subItems = sec.items.filter(subItem => hasReadableModuleAccess(subItem, permissions));

                if (subItems.length === 0) {
                    return [];
                }

                return [
                    {
                        ...sec,
                        items: subItems,
                    },
                ];
            });

            if (filteredSections.length === 0) {
                return [];
            }

            return [
                {
                    ...item,
                    sections: filteredSections,
                },
            ];
        }

        const subItems = item.items?.filter(subItem => hasReadableModuleAccess(subItem, permissions));

        if (item.items && (!subItems || subItems.length === 0)) {
            return [];
        }

        return [
            {
                ...item,
                items: subItems,
            },
        ];
    });
}

export function ModuleSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
    const { profile } = useProfileContext();
    const { map: modulePermissionMap } = useConditionalModuleCollections();
    const { isMobile, setOpenMobile } = useSidebar();
    const location = useLocation();

    React.useEffect(() => {
        if (isMobile) {
            setOpenMobile(false);
        }
    }, [location.pathname, isMobile, setOpenMobile]);

    const navigationItems = React.useMemo(
        () => filterSidebarItems(navMain, modulePermissionMap),
        [modulePermissionMap],
    );

    if (!profile) {
        return null;
    }
    return (
        <Sidebar
            collapsible="icon"
            {...props}
        >
            <SidebarHeader className="items-center justify-center p-2">
                <LogoIcon className="x-logo h-12 w-12 text-foreground" />
            </SidebarHeader>
            <SidebarContent>
                <NavMain items={navigationItems} />
            </SidebarContent>
            <SidebarFooter>
                <NavUser user={profile} />
            </SidebarFooter>
            <SidebarRail />
        </Sidebar>
    );
}
