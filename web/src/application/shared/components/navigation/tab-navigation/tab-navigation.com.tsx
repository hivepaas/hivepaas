import { type ReactNode, memo } from "react";

import { cn } from "@/lib/utils";
import { matchPath, useLocation } from "react-router";

import { AppLink } from "..";

function normalizePath(path: string) {
    return path.replace(/\/+$/, "") || "/";
}

function normalizePattern(pattern: string) {
    const normalized = normalizePath(pattern);

    return normalized.startsWith("/") ? normalized : `/${normalized}`;
}

function isExactPathMatch(route: string, pathname: string) {
    return normalizePath(route) === normalizePath(pathname);
}

function isPrefixPathMatch(prefix: string, pathname: string) {
    const normalizedPathname = normalizePath(pathname);
    const normalizedPrefix = normalizePath(prefix);

    return normalizedPathname === normalizedPrefix || normalizedPathname.startsWith(`${normalizedPrefix}/`);
}

function isPatternMatch(pattern: string, pathname: string) {
    return matchPath({ path: normalizePattern(pattern), end: false }, normalizePath(pathname)) !== null;
}

function isGroupedPathMatch(link: TabItem, pathname: string) {
    return (
        link.activePathPrefixes?.some(prefix => isPrefixPathMatch(prefix, pathname)) === true ||
        link.activePatterns?.some(pattern => isPatternMatch(pattern, pathname)) === true
    );
}

function findActiveRoute(links: TabItem[], pathname: string) {
    const exactMatch = links.find(link => isExactPathMatch(link.route, pathname));

    if (exactMatch) {
        return exactMatch.route;
    }

    return links.find(link => isGroupedPathMatch(link, pathname))?.route ?? null;
}

function tabClassName(isActive: boolean, isDisabled: boolean | undefined) {
    return cn(
        "inline-flex h-full flex-1 items-center justify-center gap-1.5 whitespace-nowrap rounded-none border-0 border-b-2 border-transparent bg-background p-3 text-md font-medium text-foreground transition-[color,box-shadow]",
        "focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:outline-ring focus-visible:ring-[3px] focus-visible:outline-1",
        isActive && "border-primary text-primary dark:border-primary",
        isDisabled && "pointer-events-none opacity-50",
    );
}

function TabLink({ link, isActive }: { link: TabItem; isActive: boolean }) {
    const className = tabClassName(isActive, link.disabled);

    if (link.disabled) {
        return (
            <span
                aria-disabled="true"
                className={className}
            >
                {link.label}
            </span>
        );
    }

    return (
        <AppLink.Modules
            to={link.route}
            ignorePrevPath
            aria-current={isActive ? "page" : undefined}
            className={className}
        >
            {link.label}
        </AppLink.Modules>
    );
}

function View({ links }: Props) {
    const location = useLocation();

    const activeRoute = findActiveRoute(links, location.pathname);

    return (
        <div className="w-full max-w-md">
            <nav
                aria-label="Tabs"
                className="gap-4"
            >
                <div className="bg-background inline-flex h-12 w-fit items-center justify-center rounded-none border-b p-0">
                    {links.map(link => (
                        <TabLink
                            key={link.route}
                            link={link}
                            isActive={activeRoute === link.route}
                        />
                    ))}
                </div>
            </nav>
        </div>
    );
}

interface TabItem {
    route: string;
    label: string | ReactNode;
    disabled?: boolean;
    activePathPrefixes?: string[];
    activePatterns?: string[];
}

interface Props {
    links: TabItem[];
}

export const TabNavigation = memo(View);
