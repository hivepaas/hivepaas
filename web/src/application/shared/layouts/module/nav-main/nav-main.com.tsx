import classnames from "classnames/bind";

import { ChevronRight, type LucideIcon } from "lucide-react";
import { matchPath, useLocation } from "react-router";

import { AppNavLink } from "@application/shared/components/navigation";

import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Separator } from "@/components/ui/separator";
import {
    SidebarGroup,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarMenuSub,
    SidebarMenuSubButton,
    SidebarMenuSubItem,
    useSidebar,
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
    return (!item.items || item.items.length === 0) && (!item.sections || item.sections.length === 0);
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
    return items.flatMap(item => {
        if (item.sections && item.sections.length > 0) {
            return item.sections.flatMap(section => getNavigationLeaves(section.items));
        }
        if (item.items && item.items.length > 0) {
            return getNavigationLeaves(item.items);
        }
        return [item];
    });
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

function NavigationLink({ route, label, Icon, isActive, isTopLevel = false }: NavigationLinkProps) {
    const { isMobile, setOpenMobile } = useSidebar();

    return (
        <AppNavLink.Modules
            to={route}
            aria-current={isActive ? "page" : undefined}
            className={cx("link")}
            onClick={() => {
                if (isMobile) {
                    setOpenMobile(false);
                }
            }}
        >
            {({ isPending }) => {
                return (
                    <SidebarMenuButton
                        // asChild
                        tooltip={label}
                        className={cx("link-content", {
                            "is-active": isActive,
                            "is-pending": isPending,
                            "is-top-level": isTopLevel,
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
    isTopLevel?: boolean;
}

interface NavigationGroupProps {
    activeKey: string | null;
    item: NavigationItem;
    pathname: string;
}

function NavigationGroup({ activeKey, item, pathname }: NavigationGroupProps) {
    const leaves = getNavigationLeaves([item]);
    const isOpen =
        leaves.some(subItem => getNavigationItemKey(subItem) === activeKey) || isPatternMatch(item.pattern, pathname);

    const { sections } = item;

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
                        className="px-4 font-medium"
                    >
                        {item.icon && <item.icon />}
                        <span>{item.title}</span>
                        <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                    </SidebarMenuButton>
                </CollapsibleTrigger>
                <CollapsibleContent>
                    <SidebarMenuSub className={sections ? "px-1.5" : undefined}>
                        {sections
                            ? sections.map((section, sIndex) => (
                                  <li
                                      key={section.title}
                                      className="flex flex-col list-none"
                                  >
                                      <div className="px-2 pt-2.5 pb-1 text-[10px] font-semibold tracking-wider text-muted-foreground/70 uppercase select-none">
                                          {section.title}
                                      </div>
                                      <ul className="flex flex-col gap-1 list-none p-0 m-0">
                                          {section.items.map(subItem => (
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
                                      </ul>
                                      {sIndex < sections.length - 1 && (
                                          <div className="py-1 px-2">
                                              <Separator className="opacity-40" />
                                          </div>
                                      )}
                                  </li>
                              ))
                            : item.items?.map(subItem => (
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

export interface NavigationSection {
    title: string;
    items: NavigationItem[];
}

export interface NavigationItem {
    title: string;
    route: string;
    pattern: string;
    icon?: LucideIcon;
    items?: NavigationItem[];
    sections?: NavigationSection[];
}

export function NavMain({ items }: { items: NavigationItem[] }) {
    const location = useLocation();
    const activeKey = findActiveNavigationKey(items, location.pathname);

    return (
        <SidebarGroup>
            {/* <SidebarGroupLabel>Platform</SidebarGroupLabel> */}
            <SidebarMenu>
                {items.map(item =>
                    (item.items && item.items.length > 0) || (item.sections && item.sections.length > 0) ? (
                        <NavigationGroup
                            key={item.title}
                            activeKey={activeKey}
                            item={item}
                            pathname={location.pathname}
                        />
                    ) : (
                        <SidebarMenuItem key={item.title}>
                            <NavigationLink
                                route={item.route}
                                label={item.title}
                                Icon={item.icon}
                                isActive={getNavigationItemKey(item) === activeKey}
                                isTopLevel
                            />
                        </SidebarMenuItem>
                    ),
                )}
            </SidebarMenu>
        </SidebarGroup>
    );
}
