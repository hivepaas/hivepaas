import { useMemo, useState } from "react";

import type { AppLogHistoryInfo, AppLogHistoryReason } from "~/projects/api/services";
import { AppLogsQueries } from "~/projects/data";

import { LogsViewer, type LogsViewerFrame } from "@application/shared/components";

import { Button, Input, Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui";

const PAGE_SIZE = 500;

const RANGES = {
    "15m": 15 * 60 * 1000,
    "1h": 60 * 60 * 1000,
    "6h": 6 * 60 * 60 * 1000,
    "24h": 24 * 60 * 60 * 1000,
    "7d": 7 * 24 * 60 * 60 * 1000,
} as const;
type Range = keyof typeof RANGES;

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
    projectID,
    env,
    appID,
    history,
    isActive,
    fontSize,
    themeId,
    height,
    isFullView,
    isFullHeight,
}: AppLogsHistoryProps) {
    const [range, setRange] = useState<Range>("1h");
    // Fixed when the range is chosen, not on every render: a start that moves
    // with the clock changes the query key each render and refetches forever.
    const [start, setStart] = useState(() => new Date(Date.now() - RANGES["1h"]));
    const [searchDraft, setSearchDraft] = useState("");
    const [search, setSearch] = useState("");
    const [levelFilter, setLevelFilter] = useState<LevelFilter>("all");
    const [stream, setStream] = useState<StreamFilter>("all");
    const isAvailable = history?.available ?? false;

    const query = AppLogsQueries.useGetHistory(
        {
            projectID,
            env,
            appID,
            start,
            limit: PAGE_SIZE,
            search: search || undefined,
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
            logViewerKey={`history:${isActive ? "active" : "inactive"}`}
            isRefreshPending={query.isFetching}
            hasLineNumbers={false}
            height={height}
            isFullView={isFullView}
            fontSize={fontSize}
            themeId={themeId}
            isFullHeight={isFullHeight}
            downloadFileName={`${appID}-history.log`}
            onRefresh={() => {
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
            toolbarFilters={
                <div className="flex min-w-0 flex-wrap items-center gap-1.5 sm:gap-2">
                    <Select
                        value={range}
                        onValueChange={value => {
                            const next = value as Range;
                            setRange(next);
                            setStart(new Date(Date.now() - RANGES[next]));
                        }}
                    >
                        <SelectTrigger className="h-8 sm:h-9 w-[90px] text-xs sm:text-sm">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            {(Object.keys(RANGES) as Range[]).map(key => (
                                <SelectItem
                                    key={key}
                                    value={key}
                                >
                                    Last {key}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                    <Select
                        value={levelFilter}
                        onValueChange={value => {
                            setLevelFilter(value as LevelFilter);
                        }}
                    >
                        <SelectTrigger className="h-8 sm:h-9 w-[170px] text-xs sm:text-sm">
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
                        <SelectTrigger className="h-8 sm:h-9 w-[110px] text-xs sm:text-sm">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">All streams</SelectItem>
                            <SelectItem value="stdout">stdout</SelectItem>
                            <SelectItem value="stderr">stderr</SelectItem>
                        </SelectContent>
                    </Select>
                    <Input
                        value={searchDraft}
                        placeholder="Search stored logs"
                        className="h-8 sm:h-9 w-[180px] text-xs sm:text-sm"
                        onChange={event => {
                            setSearchDraft(event.target.value);
                        }}
                        onKeyDown={event => {
                            if (event.key === "Enter") {
                                setSearch(searchDraft.trim());
                            }
                        }}
                    />
                </div>
            }
        />
    );
}

interface AppLogsHistoryProps {
    projectID: string;
    env: string;
    appID: string;
    history: AppLogHistoryInfo | undefined;
    isActive: boolean;
    fontSize?: number;
    themeId?: string;
    height?: number | string;
    isFullView?: boolean;
    isFullHeight?: boolean;
}
