import classnames from "classnames/bind";

import { ChevronRight, type LucideIcon } from "lucide-react";
import { matchPath, useLocation } from "react-router";

import { AppNavLink } from "@application/shared/components/navigation";

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

function normalizePath(path: string) {
    return path.replace(/\/+$/, "") || "/";
}

function normalizePattern(pattern: string) {
    const normalized = normalizePath(pattern);

    return normalized.startsWith("/") ? normalized : `/${normalized}`;
}

function isNavigationLeaf(item: NavigationItem) {
    return !item.items || item.items.length === 0;
}

function isMatchableItem(item: NavigationItem) {
    return isNavigationLeaf(item) && item.route !== "#" && item.pattern !== "#";
}

function getNavigationItemKey(item: NavigationItem) {
    return item.route;
}

function isExactPathMatch(route: string, pathname: string) {
    return route !== "#" && normalizePath(route) === normalizePath(pathname);
}

function isPatternMatch(pattern: string, pathname: string) {
    return (
        pattern !== "#" &&
        matchPath({ path: normalizePattern(pattern), caseSensitive: false, end: false }, normalizePath(pathname)) !==
            null
    );
}

function getNavigationLeaves(items: NavigationItem[]): NavigationItem[] {
    return items.flatMap(item => (item.items && item.items.length > 0 ? getNavigationLeaves(item.items) : [item]));
}

function findActiveNavigationKey(items: NavigationItem[], pathname: string) {
    const leaves = getNavigationLeaves(items).filter(isMatchableItem);
    const exactMatch = leaves.find(item => isExactPathMatch(item.route, pathname));

    if (exactMatch) {
        return getNavigationItemKey(exactMatch);
    }

    const patternMatch = leaves.find(item => isPatternMatch(item.pattern, pathname));

    return patternMatch ? getNavigationItemKey(patternMatch) : null;
}

function NavigationLink({ route, label, Icon, isActive }: NavigationLinkProps) {
    return (
        <AppNavLink.Modules
            to={route}
            aria-current={isActive ? "page" : undefined}
            className={cx("link")}
        >
            {({ isPending }) => {
                return (
                    <SidebarMenuButton
                        // asChild
                        tooltip={label}
                        className={cx("link-content", {
                            "is-active": isActive,
                            "is-pending": isPending,
                        })}
                    >
                        {Icon && <Icon className={cx("icon")} />}

                        <div className={cx("active-indicator")} />

                        <div className={cx("label")}>{label}</div>
                    </SidebarMenuButton>
                );
            }}
        </AppNavLink.Modules>
    );
}

interface NavigationLinkProps {
    route: string;
    label: string;
    Icon?: LucideIcon;
    isActive: boolean;
}

interface NavigationGroupProps {
    activeKey: string | null;
    item: NavigationItem & {
        items: NavigationItem[];
    };
    pathname: string;
}

function NavigationGroup({ activeKey, item, pathname }: NavigationGroupProps) {
    const isOpen =
        getNavigationLeaves(item.items).some(subItem => getNavigationItemKey(subItem) === activeKey) ||
        isPatternMatch(item.pattern, pathname);

    return (
        <Collapsible
            key={item.title}
            asChild
            defaultOpen={isOpen}
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
                                        label={subItem.title}
                                        Icon={subItem.icon}
                                        isActive={getNavigationItemKey(subItem) === activeKey}
                                    />
                                </SidebarMenuSubButton>
                            </SidebarMenuSubItem>
                        ))}
                    </SidebarMenuSub>
                </CollapsibleContent>
            </SidebarMenuItem>
        </Collapsible>
    );
}

interface NavigationItem {
    title: string;
    route: string;
    pattern: string;
    icon?: LucideIcon;
    items?: NavigationItem[];
}

export function NavMain({ items }: { items: NavigationItem[] }) {
    const location = useLocation();
    const activeKey = findActiveNavigationKey(items, location.pathname);

    return (
        <SidebarGroup>
            {/* <SidebarGroupLabel>Platform</SidebarGroupLabel> */}
            <SidebarMenu>
                {items.map(item =>
                    item.items && item.items.length > 0 ? (
                        <NavigationGroup
                            key={item.title}
                            activeKey={activeKey}
                            item={{ ...item, items: item.items }}
                            pathname={location.pathname}
                        />
                    ) : (
                        <SidebarMenuItem key={item.title}>
                            <NavigationLink
                                route={item.route}
                                label={item.title}
                                Icon={item.icon}
                                isActive={getNavigationItemKey(item) === activeKey}
                            />
                        </SidebarMenuItem>
                    ),
                )}
            </SidebarMenu>
        </SidebarGroup>
    );
}
