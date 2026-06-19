import { type SetStateAction, useCallback, useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import type { WebSocketReadyState } from "@infrastructure/websocket";
import { listBox } from "@lib/styles";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { AppLogsQueries } from "~/projects/data";

import { AppLink, AppLoader, type LogsViewerFrame } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui";

import { AppLogsViewer } from "../building-blocks";

const AGGREGATION_TAB_ID = "aggregation";
const DEFAULT_LOG_LINES = 100;

export function AppLogsRoute() {
    const { id: projectID, appId: appID } = useParams<{ id: string; appId: string }>();
    const [activeTab, setActiveTab] = useState(AGGREGATION_TAB_ID);
    const [tabStates, setTabStates] = useState<AppLogTabStates>(() => ({
        [AGGREGATION_TAB_ID]: createDefaultAppLogTabState(),
    }));

    invariant(projectID, "projectID must be defined");
    invariant(appID, "appID must be defined");

    const { data: infoResponse, isLoading: isInfoLoading } = AppLogsQueries.useGetInfo({
        projectID,
        appID,
    });
    const isLogsEnabled = infoResponse?.data.enabled ?? true;

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
        setTabStates(current => {
            const tabIDs = new Set(tabs.map(tab => tab.id));
            const hasMissingTabs = tabs.some(tab => current[tab.id] === undefined);
            const hasRemovedTabs = Object.keys(current).some(tabID => !tabIDs.has(tabID));

            if (!hasMissingTabs && !hasRemovedTabs) {
                return current;
            }

            return Object.fromEntries(
                tabs.map(tab => [tab.id, current[tab.id] ?? createDefaultAppLogTabState()] as const),
            );
        });
    }, [tabs]);

    const updateTabState = useCallback((tabID: string, updater: (current: AppLogTabState) => AppLogTabState) => {
        setTabStates(current => {
            const currentTabState = current[tabID] ?? createDefaultAppLogTabState();
            const nextTabState = updater(currentTabState);

            if (nextTabState === currentTabState) {
                return current;
            }

            return {
                ...current,
                [tabID]: nextTabState,
            };
        });
    }, []);

    const handleTabLogsChange = useCallback(
        (tabID: string, action: SetStateAction<LogsViewerFrame[]>) => {
            updateTabState(tabID, current => {
                const logs = resolveSetStateAction(action, current.logs);

                return logs === current.logs ? current : { ...current, logs };
            });
        },
        [updateTabState],
    );

    const handleTabLinesChange = useCallback(
        (tabID: string, action: SetStateAction<number | undefined>) => {
            updateTabState(tabID, current => {
                const lines = resolveSetStateAction(action, current.lines);

                return lines === current.lines ? current : { ...current, lines };
            });
        },
        [updateTabState],
    );

    const handleTabSinceChange = useCallback(
        (tabID: string, action: SetStateAction<Date | undefined>) => {
            updateTabState(tabID, current => {
                const since = resolveSetStateAction(action, current.since);

                return since === current.since ? current : { ...current, since };
            });
        },
        [updateTabState],
    );

    const handleTabDurationChange = useCallback(
        (tabID: string, action: SetStateAction<string | undefined>) => {
            updateTabState(tabID, current => {
                const duration = resolveSetStateAction(action, current.duration);

                return duration === current.duration ? current : { ...current, duration };
            });
        },
        [updateTabState],
    );

    const handleTabReadyStateChange = useCallback(
        (tabID: string, readyState: WebSocketReadyState) => {
            updateTabState(tabID, current => {
                return current.readyState === readyState ? current : { ...current, readyState };
            });
        },
        [updateTabState],
    );

    return (
        <section className={cn(listBox)}>
            {isInfoLoading ? (
                <AppLoader />
            ) : isLogsEnabled ? (
                <Tabs
                    value={activeTab}
                    onValueChange={setActiveTab}
                    className="gap-4"
                >
                    <TabsList className="h-auto gap-2 p-0">
                        {tabs.map(tab => {
                            const isStreaming = tabStates[tab.id]?.readyState === WebSocket.OPEN;

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

                    {tabs.map(tab => {
                        const tabState = tabStates[tab.id] ?? createDefaultAppLogTabState();

                        return (
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
                                    logs={tabState.logs}
                                    lines={tabState.lines}
                                    since={tabState.since}
                                    duration={tabState.duration}
                                    webSocketReadyState={tabState.readyState}
                                    isActive={activeTab === tab.id}
                                    shouldAutoStream={tab.id === AGGREGATION_TAB_ID}
                                    onLogsChange={handleTabLogsChange}
                                    onLinesChange={handleTabLinesChange}
                                    onSinceChange={handleTabSinceChange}
                                    onDurationChange={handleTabDurationChange}
                                    onReadyStateChange={handleTabReadyStateChange}
                                />
                            </TabsContent>
                        );
                    })}
                </Tabs>
            ) : (
                <p className="text-3xl leading-tight">
                    Logs feature is disabled, enable it in{" "}
                    <AppLink.Basic
                        to={ROUTE.projects.single.apps.single.configuration.featureSettings.$route(projectID, appID)}
                        className="text-primary underline-offset-4 hover:underline"
                    >
                        Feature Settings
                    </AppLink.Basic>
                </p>
            )}
        </section>
    );
}

function createDefaultAppLogTabState(): AppLogTabState {
    return {
        logs: [],
        lines: DEFAULT_LOG_LINES,
        since: undefined,
        duration: undefined,
        readyState: WebSocket.CLOSED,
    };
}

function resolveSetStateAction<T>(action: SetStateAction<T>, current: T): T {
    return typeof action === "function" ? (action as (current: T) => T)(current) : action;
}

interface AppLogTab {
    id: string;
    label: string;
    taskId?: string;
}

interface AppLogTabState {
    logs: LogsViewerFrame[];
    lines: number | undefined;
    since: Date | undefined;
    duration: string | undefined;
    readyState: WebSocketReadyState;
}

type AppLogTabStates = Partial<Record<string, AppLogTabState>>;
