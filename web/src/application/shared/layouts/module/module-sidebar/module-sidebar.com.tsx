import * as React from "react";

import { Container, PawPrint, Users } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";

import { Sidebar, SidebarContent, SidebarFooter, SidebarHeader, SidebarRail } from "@/components/ui/sidebar";

import { NavMain } from "../nav-main";
import { NavUser } from "../nav-user";

// This is sample data.
const data = {
    navMain: [
        {
            title: "User Management",
            route: ROUTE.userManagement.users.$route,
            pattern: ROUTE.userManagement.users.$pattern,
            icon: Users,
        },
        {
            title: "Clusters",
            route: "#",
            pattern: "#",
            icon: Container,
        },
    ],
};

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
                <NavMain items={data.navMain} />
            </SidebarContent>
            <SidebarFooter>
                <NavUser user={profile} />
            </SidebarFooter>
            <SidebarRail />
        </Sidebar>
    );
}
