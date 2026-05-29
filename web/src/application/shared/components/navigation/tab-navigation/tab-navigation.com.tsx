import { memo } from "react";

import { useLocation } from "react-router";

import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { AppLink } from "..";

function normalizePath(path: string) {
    return path.replace(/\/+$/, "");
}

function isLinkActive(link: TabItem, pathname: string) {
    const normalizedPathname = normalizePath(pathname);
    const activePaths = [link.route, ...(link.activePathPrefixes ?? [])].map(normalizePath);

    return activePaths.some(path => normalizedPathname === path || normalizedPathname.startsWith(`${path}/`));
}

function View({ links }: Props) {
    const location = useLocation();

    const activeKey = links.find(link => isLinkActive(link, location.pathname))?.route;

    return (
        <div className="w-full max-w-md">
            <Tabs
                value={activeKey}
                className="gap-4"
            >
                <TabsList className="bg-background rounded-none border-b p-0 h-12">
                    {links.map(link => (
                        <TabsTrigger
                            key={link.route}
                            value={link.route}
                            className="bg-background data-[state=active]:border-primary dark:data-[state=active]:border-primary h-full rounded-none border-0 border-b-2 border-transparent data-[state=active]:shadow-none disabled:pointer-events-none disabled:opacity-50 p-3 text-md"
                            disabled={link.disabled}
                            asChild
                        >
                            <AppLink.Modules
                                to={link.route}
                                ignorePrevPath
                            >
                                {link.label}
                            </AppLink.Modules>
                        </TabsTrigger>
                    ))}
                </TabsList>
            </Tabs>
        </div>
    );
}

interface TabItem {
    route: string;
    label: string | React.ReactNode;
    disabled?: boolean;
    activePathPrefixes?: string[];
}

interface Props {
    links: TabItem[];
}

export const TabNavigation = memo(View);
