import { type SetStateAction, useCallback, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import type { AppLogHistoryInfo, AppLogHistoryReason } from "~/projects/api/services";
import { AppLogsQueries } from "~/projects/data";

import {
    CollapsibleSearchInput,
    DEFAULT_SEARCH_MODE,
    LogsViewer,
    type LogsViewerFrame,
    type SearchMode,
    isValidRegex,
} from "@application/shared/components";

import { Button, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui";

import { AppLogsToolbarFilters } from "../app-logs-toolbar";
import { HISTORY_DURATION_OPTIONS } from "../duration-picker";
import { LOG_FILTER_FIELD_WIDTH } from "../log-filters.constants";

import {
    type LogHistoryWindow,
    describeLogHistoryView,
    earliestStoredLogTime,
    resolveDurationOptions,
    resolveLogHistoryWindow,
} from "./app-logs-history.utils";

export const DEFAULT_LOG_HISTORY_LINES = 500;

const LEVEL_FILTERS = {
    all: { label: "All levels", levels: [] },
    errors: { label: "Errors", levels: ["error", "fatal", "panic"] },
    warnings: { label: "Warnings and errors", levels: ["warn", "warning", "error", "fatal", "panic"] },
} as const;
type LevelFilter = keyof typeof LEVEL_FILTERS;

type StreamFilter = "all" | "stdout" | "stderr";

const UNAVAILABLE_TEXT: Record<AppLogHistoryReason, string> = {
    "disabled": "Stored logs are off. An administrator can turn them on in System → Logging.",
    "apps-not-collected": "App logs are not collected. An administrator can turn them on in System → Logging.",
    "no-query-endpoint": "Logs go to an external backend HivePaaS has no query endpoint for.",
    "driver-unreadable": "This app's log driver cannot be collected. Switch it to json-file in container settings.",
    "identity-missing":
        "This app was created before logging existed. Save its container settings once so its logs can be identified.",
};

export function AppLogsHistory({
    tabID,
    projectID,
    env,
    appID,
    history,
    lines,
    since,
    duration,
    isActive,
    fontSize,
    themeId,
    height,
    isFullView,
    isFullHeight,
    onLinesChange,
    onSinceChange,
    onDurationChange,
}: AppLogsHistoryProps) {
    const [searchDraft, setSearchDraft] = useState("");
    const [search, setSearch] = useState("");
    const [searchMode, setSearchMode] = useState<SearchMode>(DEFAULT_SEARCH_MODE);
    // A pattern that is not a regular expression yet would come back from the
    // server as an invalid query. Catching it here keeps the round trip, and
    // the error toast, out of someone's typing.
    const isSearchDraftInvalid = searchMode.isRegex && !isValidRegex(searchDraft);
    const [levelFilter, setLevelFilter] = useState<LevelFilter>("all");
    const [stream, setStream] = useState<StreamFilter>("all");
    const [timeWindow, reopenWindow] = useFrozenLogHistoryWindow(since, duration);
    const isAvailable = history?.available ?? false;
    const retention = history?.retention;
    const durationOptions = useMemo(() => resolveDurationOptions(HISTORY_DURATION_OPTIONS, retention), [retention]);

    const query = AppLogsQueries.useGetHistory(
        {
            projectID,
            env,
            appID,
            start: timeWindow.start,
            end: timeWindow.end,
            limit: lines && lines > 0 ? lines : DEFAULT_LOG_HISTORY_LINES,
            search: search || undefined,
            regex: searchMode.isRegex,
            matchCase: searchMode.isCaseSensitive,
            levels: [...LEVEL_FILTERS[levelFilter].levels],
            streams: stream === "all" ? [] : [stream],
        },
        { enabled: isActive && isAvailable },
    );

    // Pages arrive newest first and each holds its lines oldest first, so the
    // oldest page goes on top.
    const frames = useMemo<LogsViewerFrame[]>(
        () =>
            [...(query.data?.pages ?? [])].reverse().flatMap(page =>
                page.data.logs.map(frame => ({
                    type: frame.type as LogsViewerFrame["type"],
                    data: frame.data,
                    ts: frame.ts,
                })),
            ),
        [query.data],
    );

    if (!isAvailable) {
        return <p className="text-base">{UNAVAILABLE_TEXT[history?.reason ?? "disabled"]}</p>;
    }

    return (
        <LogsViewer
            frames={frames}
            isRefreshPending={query.isFetching}
            hasLineNumbers={false}
            height={height}
            isFullView={isFullView}
            fontSize={fontSize}
            themeId={themeId}
            isFullHeight={isFullHeight}
            downloadFileName={`${appID}-history.log`}
            status={describeLogHistoryView({
                window: timeWindow,
                lineCount: frames.length,
                oldestLoaded: frames[0]?.ts ?? undefined,
                hasMore: query.hasNextPage,
                isLoading: query.isPending,
            })}
            onRefresh={() => {
                // A window that ends at "now" was resolved once and has been
                // going stale since. Refetching it would replay the same stale
                // end; re-resolving it is what "refresh" has to mean.
                if (timeWindow.endsNow) {
                    reopenWindow();
                    return;
                }
                void query.refetch();
            }}
            toolbarStart={
                <Button
                    type="button"
                    variant="outline"
                    className="h-8 sm:h-9 text-xs sm:text-sm"
                    disabled={!query.hasNextPage || query.isFetchingNextPage}
                    isLoading={query.isFetchingNextPage}
                    onClick={() => {
                        void query.fetchNextPage();
                    }}
                >
                    Load older
                </Button>
            }
            toolbarSearch={
                <CollapsibleSearchInput
                    value={searchDraft}
                    placeholder="Query stored logs..."
                    label="Query stored logs..."
                    expandedClassName="w-48 sm:w-64 mr-1.5 sm:mr-2"
                    mode={searchMode}
                    isInvalid={isSearchDraftInvalid}
                    onValueChange={setSearchDraft}
                    onModeChange={setSearchMode}
                    onClear={() => {
                        setSearchDraft("");
                        setSearch("");
                    }}
                    onKeyDown={event => {
                        if (event.key === "Enter" && !isSearchDraftInvalid) {
                            setSearch(searchDraft.trim());
                        }
                    }}
                />
            }
            toolbarFilters={
                <div className="flex min-w-0 flex-wrap items-center gap-1.5 sm:gap-2">
                    <AppLogsToolbarFilters
                        lines={lines}
                        since={since}
                        duration={duration}
                        durationOptions={durationOptions}
                        sinceFromDate={earliestStoredLogTime(retention, new Date())}
                        onLinesChange={value => {
                            onLinesChange(tabID, value);
                        }}
                        onSinceChange={value => {
                            onSinceChange(tabID, value);
                        }}
                        onDurationChange={value => {
                            onDurationChange(tabID, value);
                        }}
                        // Changing the window changes the query key, which
                        // refetches on its own - there is nothing to nudge.
                        onRequestRefresh={() => undefined}
                    />
                    <Select
                        value={levelFilter}
                        onValueChange={value => {
                            setLevelFilter(value as LevelFilter);
                        }}
                    >
                        <SelectTrigger className={cn("h-8 sm:h-9 text-xs sm:text-sm", LOG_FILTER_FIELD_WIDTH)}>
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            {(Object.keys(LEVEL_FILTERS) as LevelFilter[]).map(key => (
                                <SelectItem
                                    key={key}
                                    value={key}
                                >
                                    {LEVEL_FILTERS[key].label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                    <Select
                        value={stream}
                        onValueChange={value => {
                            setStream(value as StreamFilter);
                        }}
                    >
                        <SelectTrigger className={cn("h-8 sm:h-9 text-xs sm:text-sm", LOG_FILTER_FIELD_WIDTH)}>
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">All streams</SelectItem>
                            <SelectItem value="stdout">stdout</SelectItem>
                            <SelectItem value="stderr">stderr</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
            }
        />
    );
}

/**
 * useFrozenLogHistoryWindow resolves Since and Duration once per change.
 *
 * "Last 7d" has to mean seven days from one fixed instant. Recomputing it every
 * render would move `start`, and a moving `start` is a new query key on every
 * render - the query would refetch forever.
 */
function useFrozenLogHistoryWindow(
    since: Date | undefined,
    duration: string | undefined,
): [LogHistoryWindow, () => void] {
    const [resolved, setResolved] = useState(() => ({
        since,
        duration,
        window: resolveLogHistoryWindow(since, duration, new Date()),
    }));

    if (resolved.since !== since || resolved.duration !== duration) {
        setResolved({ since, duration, window: resolveLogHistoryWindow(since, duration, new Date()) });
    }

    // Resolving against a fresh clock changes the query key, so the pages start
    // over from the newest line rather than from where the old window ended.
    const reopen = useCallback(() => {
        setResolved(current => ({
            ...current,
            window: resolveLogHistoryWindow(current.since, current.duration, new Date()),
        }));
    }, []);

    return [resolved.window, reopen];
}

interface AppLogsHistoryProps {
    tabID: string;
    projectID: string;
    env: string;
    appID: string;
    history: AppLogHistoryInfo | undefined;
    lines: number | undefined;
    since: Date | undefined;
    duration: string | undefined;
    isActive: boolean;
    fontSize?: number;
    themeId?: string;
    height?: number | string;
    isFullView?: boolean;
    isFullHeight?: boolean;
    onLinesChange: (tabID: string, action: SetStateAction<number | undefined>) => void;
    onSinceChange: (tabID: string, action: SetStateAction<Date | undefined>) => void;
    onDurationChange: (tabID: string, action: SetStateAction<string | undefined>) => void;
}
