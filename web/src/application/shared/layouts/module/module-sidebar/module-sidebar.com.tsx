import * as React from "react";

import { Container, LayoutGrid, type LucideIcon, PawPrint, Users } from "lucide-react";

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
        title: "User Management",
        route: ROUTE.userManagement.users.$route,
        pattern: ROUTE.userManagement.users.$pattern,
        icon: Users,
    },
    {
        title: "Projects",
        route: ROUTE.projects.list.$route,
        pattern: ROUTE.projects.list.$pattern,
        icon: LayoutGrid,
    },
    {
        title: "Clusters",
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
