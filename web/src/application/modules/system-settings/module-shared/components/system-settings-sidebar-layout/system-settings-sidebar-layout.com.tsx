import { type PropsWithChildren, memo, useMemo } from "react";

import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import type { LucideIcon } from "lucide-react";
import { useLocation, useNavigate } from "react-router";

import { AppLink } from "@application/shared/components";

import {
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectLabel,
    SelectSeparator,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

export interface SystemSettingsTabItem {
    route: string;
    label: string;
    icon: LucideIcon;
    disabled?: boolean;
}

export interface SystemSettingsTabSection {
    title: string;
    items: SystemSettingsTabItem[];
}

export interface SystemSettingsSidebarLayoutProps extends PropsWithChildren {
    sections: SystemSettingsTabSection[];
}

function normalizePath(path: string) {
    return path.replace(/\/+$/, "");
}

function findActiveTab(tabs: { route: string }[], pathname: string) {
    const normalizedPathname = normalizePath(pathname);
    const exactMatch = tabs.find(({ route }) => normalizePath(route) === normalizedPathname);
    if (exactMatch) {
        return exactMatch.route;
    }
    return tabs.find(({ route }) => normalizedPathname.startsWith(`${normalizePath(route)}/`))?.route;
}

function View({ sections, children }: SystemSettingsSidebarLayoutProps) {
    const location = useLocation();
    const navigate = useNavigate();

    const allTabs = useMemo(() => sections.flatMap(sec => sec.items), [sections]);
    const activeKey = findActiveTab(allTabs, location.pathname);

    return (
        <div className="flex flex-col gap-2 md:flex-row md:gap-4 w-full max-w-[1400px] mx-auto min-w-0">
            {/* Mobile Dropdown Navigation (< md) */}
            <div className="md:hidden w-full bg-background/95 backdrop-blur-md rounded-xl p-2 shadow-xs sticky top-[53px] z-20 border border-amber-500/25">
                <Select
                    value={activeKey}
                    onValueChange={val => {
                        if (val) {
                            void navigate(val);
                        }
                    }}
                >
                    <SelectTrigger className="w-full bg-gradient-to-r from-amber-500/10 via-amber-400/15 to-yellow-500/10 hover:from-amber-500/20 hover:to-yellow-500/20 border-amber-500/40 dark:border-amber-400/40 text-foreground font-semibold shadow-xs h-10 px-3.5 rounded-lg focus:ring-amber-400/40">
                        <SelectValue placeholder="Select section..." />
                    </SelectTrigger>
                    <SelectContent>
                        {sections.map((sec, sIndex) => (
                            <SelectGroup key={sec.title}>
                                <SelectLabel className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground/70 px-2 py-1.5">
                                    {sec.title}
                                </SelectLabel>
                                {sec.items.map(tab => {
                                    const Icon = tab.icon;
                                    return (
                                        <SelectItem
                                            key={tab.route}
                                            value={tab.route}
                                            disabled={tab.disabled}
                                            className="cursor-pointer"
                                        >
                                            <div className="flex items-center gap-2">
                                                <Icon className="size-4 shrink-0 text-muted-foreground" />
                                                <span>{tab.label}</span>
                                            </div>
                                        </SelectItem>
                                    );
                                })}
                                {sIndex < sections.length - 1 && <SelectSeparator />}
                            </SelectGroup>
                        ))}
                    </SelectContent>
                </Select>
            </div>

            {/* Desktop Vertical Sidebar Navigation (>= md) */}
            <aside className="hidden md:block w-56 shrink-0">
                <div className="sticky top-4">
                    <div className="bg-background rounded-lg py-3 shadow-xs">
                        <Tabs
                            value={activeKey}
                            className="flex-row"
                        >
                            <TabsList className="bg-background h-full flex-col w-full rounded-none p-0 gap-3">
                                {sections.map((sec, sIndex) => (
                                    <div
                                        key={sec.title}
                                        className="w-full flex flex-col"
                                    >
                                        <div className="px-3.5 pb-1.5 text-[11px] font-semibold tracking-wider text-muted-foreground/70 uppercase select-none">
                                            {sec.title}
                                        </div>
                                        <div className="flex flex-col">
                                            {sec.items.map(tab => {
                                                const Icon = tab.icon;
                                                return (
                                                    <TabsTrigger
                                                        key={tab.route}
                                                        value={tab.route}
                                                        asChild
                                                        disabled={tab.disabled}
                                                        className="py-2 px-3.5 cursor-pointer bg-background font-medium h-auto w-full justify-start rounded-none border-0 border-l-2 border-transparent transition-colors text-foreground hover:bg-muted/40 data-[state=active]:border-amber-500 dark:data-[state=active]:border-amber-400 data-[state=active]:bg-amber-500/[0.08] dark:data-[state=active]:bg-amber-400/[0.1] data-[state=active]:shadow-none [&[data-state=active]_.tab-icon]:text-amber-500 dark:[&[data-state=active]_.tab-icon]:text-amber-400"
                                                    >
                                                        <AppLink.Basic
                                                            className="w-full flex items-center gap-2.5 text-left text-sm"
                                                            to={tab.route}
                                                            ignorePrevPath
                                                        >
                                                            <Icon className="tab-icon size-4 shrink-0 transition-colors text-muted-foreground" />
                                                            <span className="truncate">{tab.label}</span>
                                                        </AppLink.Basic>
                                                    </TabsTrigger>
                                                );
                                            })}
                                        </div>
                                        {sIndex < sections.length - 1 && (
                                            <div className="pt-2 px-3.5">
                                                <Separator className="opacity-40" />
                                            </div>
                                        )}
                                    </div>
                                ))}
                            </TabsList>
                        </Tabs>
                    </div>
                </div>
            </aside>
            <div className={cn(listBox, "w-full min-w-0 flex-1")}>{children}</div>
        </div>
    );
}

export const SystemSettingsSidebarLayout = memo(View);
