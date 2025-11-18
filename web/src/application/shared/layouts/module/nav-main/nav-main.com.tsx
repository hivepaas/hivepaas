import classnames from "classnames/bind";

import { ChevronRight, type LucideIcon } from "lucide-react";

import { AppNavLink } from "@application/shared/components/navigation";
import { useAppMatchPath } from "@application/shared/hooks/router";

import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import {
    SidebarGroup,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarMenuSub,
    SidebarMenuSubButton,
    SidebarMenuSubItem,
} from "@/components/ui/sidebar";

import styles from "./nav-main.module.scss";

const cx = classnames.bind(styles);

function NavigationLink({ route, pattern, label, Icon }: NavigationLinkProps) {
    const { match } = useAppMatchPath();

    const isActive = match.modules({ pattern }) !== null;

    return (
        <AppNavLink.Modules
            to={route}
            className={cx("link")}
        >
            {({ isPending }) => {
                return (
                    <div
                        className={cx("link-content", {
                            "is-active": isActive,
                            "is-pending": isPending,
                        })}
                    >
                        <div className={cx("active-indicator")} />

                        {Icon && <Icon className={cx("icon")} />}
                        <div className={cx("label")}>{label}</div>
                    </div>
                );
            }}
        </AppNavLink.Modules>
    );
}

interface NavigationLinkProps {
    route: string;
    pattern: string;
    label: string;
    Icon?: LucideIcon;
}

export function NavMain({
    items,
}: {
    items: {
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
    }[];
}) {
    return (
        <SidebarGroup>
            {/* <SidebarGroupLabel>Platform</SidebarGroupLabel> */}
            <SidebarMenu>
                {items.map(item =>
                    item.items && item.items.length > 0 ? (
                        <Collapsible
                            key={item.title}
                            asChild
                            className="group/collapsible"
                        >
                            <SidebarMenuItem>
                                <CollapsibleTrigger asChild>
                                    <SidebarMenuButton
                                        tooltip={item.title}
                                        className="px-4"
                                    >
                                        {item.icon && <item.icon />}
                                        <span>{item.title}</span>
                                        <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                                    </SidebarMenuButton>
                                </CollapsibleTrigger>
                                <CollapsibleContent>
                                    <SidebarMenuSub>
                                        {item.items.map(subItem => (
                                            <SidebarMenuSubItem key={subItem.title}>
                                                <SidebarMenuSubButton asChild>
                                                    <NavigationLink
                                                        route={subItem.route}
                                                        pattern={subItem.pattern}
                                                        label={subItem.title}
                                                        Icon={subItem.icon}
                                                    />
                                                </SidebarMenuSubButton>
                                            </SidebarMenuSubItem>
                                        ))}
                                    </SidebarMenuSub>
                                </CollapsibleContent>
                            </SidebarMenuItem>
                        </Collapsible>
                    ) : (
                        <SidebarMenuItem key={item.title}>
                            <SidebarMenuButton
                                asChild
                                tooltip={item.title}
                            >
                                <NavigationLink
                                    route={item.route}
                                    pattern={item.pattern}
                                    label={item.title}
                                    Icon={item.icon}
                                />
                            </SidebarMenuButton>
                        </SidebarMenuItem>
                    ),
                )}
            </SidebarMenu>
        </SidebarGroup>
    );
}
