import { useCallback, useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import type { WebSocketReadyState } from "@infrastructure/websocket";
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
    const [tabReadyStates, setTabReadyStates] = useState<Record<string, WebSocketReadyState>>({});

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

    useEffect(() => {
        setTabReadyStates(current => {
            const next = Object.fromEntries(
                Object.entries(current).filter(([tabID]) => tabs.some(tab => tab.id === tabID)),
            );

            return Object.keys(next).length === Object.keys(current).length ? current : next;
        });
    }, [tabs]);

    const handleTabReadyStateChange = useCallback((tabID: string, readyState: WebSocketReadyState) => {
        setTabReadyStates(current => (current[tabID] === readyState ? current : { ...current, [tabID]: readyState }));
    }, []);

    return (
        <section className={cn(listBox)}>
            <Tabs
                value={activeTab}
                onValueChange={setActiveTab}
                className="gap-4"
            >
                <TabsList className="h-auto gap-2 p-0">
                    {tabs.map(tab => {
                        const isStreaming = tabReadyStates[tab.id] === WebSocket.OPEN;

                        return (
                            <TabsTrigger
                                key={tab.id}
                                value={tab.id}
                            >
                                <span>{tab.label}</span>
                                {isStreaming && (
                                    <span
                                        aria-label={`${tab.label} streaming`}
                                        className="size-2.5 rounded-full bg-rose-500"
                                    />
                                )}
                            </TabsTrigger>
                        );
                    })}
                </TabsList>

                {tabs.map(tab => (
                    <TabsContent
                        key={tab.id}
                        value={tab.id}
                        forceMount
                        className="m-0 data-[state=inactive]:hidden"
                    >
                        <AppLogsViewer
                            tabID={tab.id}
                            projectID={projectID}
                            appID={appID}
                            tabLabel={tab.label}
                            taskId={tab.taskId}
                            isActive={activeTab === tab.id}
                            shouldAutoStream={tab.id === AGGREGATION_TAB_ID}
                            onReadyStateChange={handleTabReadyStateChange}
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
