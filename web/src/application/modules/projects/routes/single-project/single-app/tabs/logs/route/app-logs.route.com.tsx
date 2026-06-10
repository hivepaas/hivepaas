import { useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { AppLogsQueries } from "~/projects/data";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui";

import { AppLogsViewer } from "../building-blocks";

const AGGREGATION_TAB_ID = "aggregation";

export function AppLogsRoute() {
    const { id: projectID, appId: appID } = useParams<{ id: string; appId: string }>();
    const [activeTab, setActiveTab] = useState(AGGREGATION_TAB_ID);

    invariant(projectID, "projectID must be defined");
    invariant(appID, "appID must be defined");

    const { data: infoResponse } = AppLogsQueries.useGetInfo({
        projectID,
        appID,
    });

    const tabs = useMemo<AppLogTab[]>(
        () => [
            {
                id: AGGREGATION_TAB_ID,
                label: "Aggregation",
            },
            ...(infoResponse?.data.tasks ?? []).map((task, index) => ({
                id: task.id,
                label: `I${index + 1}`,
                taskId: task.id,
            })),
        ],
        [infoResponse?.data.tasks],
    );

    useEffect(() => {
        if (!tabs.some(tab => tab.id === activeTab)) {
            setActiveTab(AGGREGATION_TAB_ID);
        }
    }, [activeTab, tabs]);

    return (
        <section className={cn(listBox)}>
            <Tabs
                value={activeTab}
                onValueChange={setActiveTab}
                className="gap-4"
            >
                <TabsList className="h-auto gap-2 p-0">
                    {tabs.map(tab => (
                        <TabsTrigger
                            key={tab.id}
                            value={tab.id}
                        >
                            {tab.label}
                        </TabsTrigger>
                    ))}
                </TabsList>

                {tabs.map(tab => (
                    <TabsContent
                        key={tab.id}
                        value={tab.id}
                        forceMount
                        className="m-0 data-[state=inactive]:hidden"
                    >
                        <AppLogsViewer
                            projectID={projectID}
                            appID={appID}
                            tabLabel={tab.label}
                            taskId={tab.taskId}
                            isActive={activeTab === tab.id}
                        />
                    </TabsContent>
                ))}
            </Tabs>
        </section>
    );
}

interface AppLogTab {
    id: string;
    label: string;
    taskId?: string;
}
