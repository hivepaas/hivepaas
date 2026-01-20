import { memo, type PropsWithChildren } from "react";

import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ROUTE } from "@application/shared/constants";
import { useLocation } from "react-router";
import { AppLink } from "@application/shared/components";



interface TabItem {
    route: string;
    label: string | React.ReactNode;
    disabled?: boolean;
}

function View({ projectId, children }: Props) {
    const location = useLocation();
    const tabs: TabItem[] = [
        {
            label: "General",
            route: ROUTE.projects.single.configuration.general.$route(projectId),
        },
        {
            label: "Env Variables",
            route: ROUTE.projects.single.configuration.envVariables.$route(projectId),
        },
        {
            label: "Secrets",
            route: ROUTE.projects.single.configuration.secrets.$route(projectId),
        },
    ];

    const activeKey = tabs.find(({ route }) => route === location.pathname)?.route;
    return (
        <div className="flex flex-col gap-5 md:flex-row max-w-7xl w-full mx-auto">

            <aside className="md:w-56 md:shrink-0">
                <div className="sticky top-4">
                    <div className="bg-background rounded-lg py-4">
                        <Tabs
                            defaultValue={activeKey}
                            className="flex-row"
                        >
                            <TabsList className="bg-background h-full flex-col w-full rounded-none p-0">
                                {tabs.map(tab => (
                                    <TabsTrigger
                                        key={tab.route}
                                        value={tab.route}
                                        className="py-3 cursor-pointer bg-background data-[state=active]:border-primary data-[state=active]:bg-primary/10 dark:data-[state=active]:border-primary h-full w-full justify-start rounded-none border-0 border-l-2 border-transparent data-[state=active]:shadow-none"
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
            </aside >
            <div className="flex-1 min-w-0 bg-background rounded-lg p-4 max-w-7xl w-full">

                {children}
            </div>
        </div>
    );
}

interface Props extends PropsWithChildren {
    projectId: string;
}

export const ProjectWithSidebar = memo(View);
