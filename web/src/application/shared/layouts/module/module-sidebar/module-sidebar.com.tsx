import * as React from "react";

import {
    Container,
    LayoutGrid,
    type LucideIcon,
    PawPrint,
    Settings,
    Settings2,
    SlidersHorizontal,
    Users,
} from "lucide-react";

import { MODULE_IDS, ROUTE, type ResourceModuleId } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";
import { type ModuleId, type ModulePermission, useConditionalModuleCollections } from "@application/shared/permissions";

import { Sidebar, SidebarContent, SidebarFooter, SidebarHeader, SidebarRail } from "@/components/ui/sidebar";

import { NavMain } from "../nav-main";
import { NavUser } from "../nav-user";

interface SidebarItem {
    title: string;
    route: string;
    pattern: string;
    icon?: LucideIcon;
    moduleId?: ResourceModuleId;
    alwaysVisible?: boolean;
    items?: {
        title: string;
        route: string;
        pattern: string;
        icon?: LucideIcon;
        moduleId?: ResourceModuleId;
    }[];
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
        title: "Sources",
        route: "#",
        pattern: ROUTE.sources.$pattern,
        icon: SlidersHorizontal,
        moduleId: MODULE_IDS.Settings,
        items: [
            {
                title: "Github Apps",
                route: ROUTE.sources.githubApps.$route,
                pattern: ROUTE.sources.githubApps.$pattern,
            },
            {
                title: "Webhooks",
                route: ROUTE.sources.webhooks.$route,
                pattern: ROUTE.sources.webhooks.$pattern,
            },
        ],
    },
    {
        title: "Settings",
        route: "#",
        pattern: ROUTE.settings.$pattern,
        icon: Settings,
        moduleId: MODULE_IDS.Settings,
        items: [
            {
                title: "Access Tokens",
                route: ROUTE.settings.accessTokens.$route,
                pattern: ROUTE.settings.accessTokens.$pattern,
            },
            {
                title: "Basic Auth",
                route: ROUTE.settings.basicAuth.$route,
                pattern: ROUTE.settings.basicAuth.$pattern,
            },
            {
                title: "Cloud Storages",
                route: ROUTE.settings.cloudStorages.$route,
                pattern: ROUTE.settings.cloudStorages.$pattern,
            },
            {
                title: "Email Accounts",
                route: ROUTE.settings.emailAccounts.$route,
                pattern: ROUTE.settings.emailAccounts.$pattern,
            },
            {
                title: "IM Platforms",
                route: ROUTE.settings.imPlatforms.$route,
                pattern: ROUTE.settings.imPlatforms.$pattern,
            },
            {
                title: "Notification Targets",
                route: ROUTE.settings.notificationTargets.$route,
                pattern: ROUTE.settings.notificationTargets.$pattern,
            },
            {
                title: "OAuth",
                route: ROUTE.settings.oauth.$route,
                pattern: ROUTE.settings.oauth.$pattern,
            },
            {
                title: "Registry Auth",
                route: ROUTE.settings.registryAuth.$route,
                pattern: ROUTE.settings.registryAuth.$pattern,
            },
            {
                title: "SSH Keys",
                route: ROUTE.settings.sshKeys.$route,
                pattern: ROUTE.settings.sshKeys.$pattern,
            },
            {
                title: "SSL Certificates",
                route: ROUTE.settings.sslCertificates.$route,
                pattern: ROUTE.settings.sslCertificates.$pattern,
            },
        ],
    },
    {
        title: "System Settings",
        route: "#",
        pattern: ROUTE.systemSettings.$pattern,
        icon: Settings2,
        moduleId: MODULE_IDS.System,
        items: [
            {
                title: "Data Backup",
                route: ROUTE.systemSettings.dataBackup.configuration.$route,
                pattern: ROUTE.systemSettings.dataBackup.$pattern,
            },
            {
                title: "Data Cleanup",
                route: ROUTE.systemSettings.dataCleanup.configuration.$route,
                pattern: ROUTE.systemSettings.dataCleanup.$pattern,
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
                route: ROUTE.cluster.nodes.$route,
                pattern: ROUTE.cluster.nodes.$pattern,
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
];

function hasReadableModuleAccess(item: SidebarItem, permissions: ReadonlyMap<ModuleId, ModulePermission>) {
    return (item.alwaysVisible ?? false) || !item.moduleId || permissions.get(item.moduleId)?.actions.read === true;
}

function filterSidebarItems(items: readonly SidebarItem[], permissions: ReadonlyMap<ModuleId, ModulePermission>) {
    return items.flatMap(item => {
        if (!hasReadableModuleAccess(item, permissions)) {
            return [];
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
                <PawPrint size={36} />
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
