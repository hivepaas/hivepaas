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

import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";

import { Sidebar, SidebarContent, SidebarFooter, SidebarHeader, SidebarRail } from "@/components/ui/sidebar";

import { NavMain } from "../nav-main";
import { NavUser } from "../nav-user";

interface SidebarItem {
    title: string;
    route: string;
    pattern: string;
    icon?: LucideIcon;
    items?: {
        title: string;
        route: string;
        pattern: string;
        icon?: LucideIcon;
    }[];
}

const navMain: SidebarItem[] = [
    {
        title: "Projects",
        route: ROUTE.projects.list.$route,
        pattern: ROUTE.projects.list.$pattern,
        icon: LayoutGrid,
    },
    {
        title: "Cluster",
        route: "#",
        pattern: "#",
        icon: Container,
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
    },
    {
        title: "Sources",
        route: "#",
        pattern: ROUTE.sources.$pattern,
        icon: SlidersHorizontal,
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
        items: [
            {
                title: "Basic Auth",
                route: ROUTE.settings.basicAuth.$route,
                pattern: ROUTE.settings.basicAuth.$pattern,
            },
            {
                title: "Registry Auth",
                route: ROUTE.settings.registryAuth.$route,
                pattern: ROUTE.settings.registryAuth.$pattern,
            },
            {
                title: "SSL Certificates",
                route: ROUTE.settings.sslCertificates.$route,
                pattern: ROUTE.settings.sslCertificates.$pattern,
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
                title: "SSH Keys",
                route: ROUTE.settings.sshKeys.$route,
                pattern: ROUTE.settings.sshKeys.$pattern,
            },
            {
                title: "Access Tokens",
                route: ROUTE.settings.accessTokens.$route,
                pattern: ROUTE.settings.accessTokens.$pattern,
            },
            {
                title: "Cloud Storages",
                route: ROUTE.settings.cloudStorages.$route,
                pattern: ROUTE.settings.cloudStorages.$pattern,
            },
            {
                title: "OAuth",
                route: ROUTE.settings.oauth.$route,
                pattern: ROUTE.settings.oauth.$pattern,
            },
            {
                title: "Notification Targets",
                route: ROUTE.settings.notificationTargets.$route,
                pattern: ROUTE.settings.notificationTargets.$pattern,
            },
        ],
    },
    {
        title: "System Settings",
        route: "#",
        pattern: ROUTE.systemSettings.$pattern,
        icon: Settings2,
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
];

export function ModuleSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
    const { profile } = useProfileContext();

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
                <NavMain items={navMain} />
            </SidebarContent>
            <SidebarFooter>
                <NavUser user={profile} />
            </SidebarFooter>
            <SidebarRail />
        </Sidebar>
    );
}
