import { type PropsWithChildren, memo } from "react";

import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useLocation, useNavigate } from "react-router";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

interface TabItem {
    route: string;
    label: string;
}

function normalizePath(path: string) {
    return path.replace(/\/+$/, "");
}

function findActiveTab(tabs: TabItem[], pathname: string) {
    const normalizedPathname = normalizePath(pathname);
    const exactMatch = tabs.find(({ route }) => normalizePath(route) === normalizedPathname);
    if (exactMatch) {
        return exactMatch.route;
    }
    return tabs.find(({ route }) => normalizedPathname.startsWith(`${normalizePath(route)}/`))?.route;
}

function View({ children }: PropsWithChildren) {
    const location = useLocation();
    const navigate = useNavigate();

    const tabs: TabItem[] = [
        {
            label: "Configuration",
            route: ROUTE.systemSettings.dataBackup.configuration.$route,
        },
        {
            label: "Backup Files",
            route: ROUTE.systemSettings.dataBackup.backupFiles.$route,
        },
        {
            label: "Actions",
            route: ROUTE.systemSettings.dataBackup.actions.$route,
        },
    ];

    const activeKey = findActiveTab(tabs, location.pathname);

    return (
        <div className="flex flex-col gap-2 md:flex-row md:gap-5 w-full max-w-[1400px] mx-auto min-w-0">
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
                        {tabs.map(tab => (
                            <SelectItem
                                key={tab.route}
                                value={tab.route}
                            >
                                {tab.label}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>

            {/* Desktop Vertical Sidebar Navigation (>= md) */}
            <aside className="hidden md:block w-56 shrink-0">
                <div className="sticky top-4">
                    <div className="bg-background rounded-lg py-4">
                        <Tabs
                            value={activeKey}
                            className="flex-row"
                        >
                            <TabsList className="bg-background h-full flex-col w-full rounded-none p-0">
                                {tabs.map(tab => (
                                    <TabsTrigger
                                        key={tab.route}
                                        value={tab.route}
                                        asChild
                                        className="py-3 pl-4 cursor-pointer bg-background data-[state=active]:border-primary data-[state=active]:bg-primary/10 dark:data-[state=active]:border-primary h-full w-full justify-start rounded-none border-0 border-l-2 border-transparent data-[state=active]:shadow-none"
                                    >
                                        <AppLink.Basic
                                            className="w-full text-left"
                                            to={tab.route}
                                            ignorePrevPath
                                        >
                                            {tab.label}
                                        </AppLink.Basic>
                                    </TabsTrigger>
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

export const DataBackupLayout = memo(View);
