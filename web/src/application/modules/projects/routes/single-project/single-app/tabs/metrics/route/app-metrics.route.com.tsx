import { useState } from "react";

import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { RefreshCw } from "lucide-react";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import type {
    AppHttpMetricsReason,
    AppLogHistoryReason,
    AppLogs_GetFunctionMetrics_Res,
    AppLogs_GetHttpMetrics_Res,
    FunctionMetricsCounts,
    FunctionMetricsRange,
    HttpMetricsCounts,
} from "~/projects/api/services";
import { AppLogsQueries, ProjectAppsQueries } from "~/projects/data";
import { isFunctionApp } from "~/projects/module-shared/utils";

import { AppLink, AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { LOG_HISTORY_UNAVAILABLE_TEXT } from "../../logs/building-blocks";
import {
    CallsChart,
    DurationChart,
    HttpPaths,
    HttpReplicas,
    METRICS_RANGES,
    MetricsPaths,
    RequestsChart,
    storeMetricsRange,
    storedMetricsRange,
} from "../building-blocks";

type MetricsView = "calls" | "http";

/**
 * An app's numbers over a range ending now: its HTTP requests, counted from
 * Traefik's access log, for every app reached by a domain; and a function's
 * calls, counted from the invocation line its runtime writes for every call.
 */
export function AppMetricsRoute() {
    const { id: projectID, env, appId: appID } = useParams<{ id: string; env: string; appId: string }>();
    invariant(projectID, "projectID must be defined");
    invariant(env, "env must be defined");
    invariant(appID, "appID must be defined");

    // The header's query: the app is already loaded.
    const { data: app } = ProjectAppsQueries.useFindOneById({ projectID, env, appID, getStats: true });
    const isFunction = app ? isFunctionApp(app.data) : false;

    const [view, setView] = useState<MetricsView>("calls");
    const activeView: MetricsView = isFunction ? view : "http";
    const [range, setRange] = useState<FunctionMetricsRange>(storedMetricsRange);

    const request = { projectID, env, appID, range };
    const callsQuery = AppLogsQueries.useGetFunctionMetrics(request, {
        enabled: Boolean(app) && activeView === "calls",
    });
    const httpQuery = AppLogsQueries.useGetHttpMetrics(request, { enabled: Boolean(app) && activeView === "http" });
    const activeQuery = activeView === "calls" ? callsQuery : httpQuery;

    function chooseRange(next: FunctionMetricsRange) {
        setRange(next);
        storeMetricsRange(next);
    }

    return (
        <div className={cn(listBox, "flex flex-col gap-4")}>
            <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex flex-wrap items-center gap-3">
                    {isFunction && (
                        <Tabs
                            value={view}
                            onValueChange={value => {
                                setView(value as MetricsView);
                            }}
                        >
                            <TabsList>
                                <TabsTrigger value="calls">Calls</TabsTrigger>
                                <TabsTrigger value="http">HTTP</TabsTrigger>
                            </TabsList>
                        </Tabs>
                    )}
                    <div className="flex gap-1">
                        {METRICS_RANGES.map(item => (
                            <Button
                                key={item.value}
                                type="button"
                                size="sm"
                                variant={item.value === range ? "default" : "outline"}
                                onClick={() => {
                                    chooseRange(item.value);
                                }}
                            >
                                {item.label}
                            </Button>
                        ))}
                    </div>
                </div>
                <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    isLoading={activeQuery.isFetching}
                    onClick={() => {
                        void activeQuery.refetch();
                    }}
                >
                    <RefreshCw /> Refresh
                </Button>
            </div>

            {!app ? (
                <AppLoader />
            ) : activeView === "calls" ? (
                <FunctionCallsView
                    metrics={callsQuery.data?.data}
                    isLoading={callsQuery.isLoading}
                    range={range}
                />
            ) : (
                <HttpRequestsView
                    metrics={httpQuery.data?.data}
                    isLoading={httpQuery.isLoading}
                    range={range}
                />
            )}
        </div>
    );
}

const HTTP_UNAVAILABLE_TEXT: Record<Exclude<AppHttpMetricsReason, AppLogHistoryReason>, string> = {
    "not-exposed":
        "The app has no domain. Its requests are counted where they enter, at Traefik, and none of its go through it.",
    "access-log-off":
        "Traefik's access log is off. An administrator turns it on in System → Traefik → Config Options, with Access Log; Traefik restarts briefly.",
    "access-log-not-json":
        "Traefik's access log is written in an older form. An administrator saves System → Traefik → Config Options once, with Access Log on; Traefik restarts briefly.",
    "access-log-unlabelled":
        "Traefik's log lines do not carry its identity yet. An administrator saves System → Traefik → Config Options once, with Access Log on; Traefik restarts briefly.",
};

function httpUnavailableText(reason: AppHttpMetricsReason | null): string {
    if (!reason) {
        return "The app's requests cannot be counted.";
    }
    if (reason in HTTP_UNAVAILABLE_TEXT) {
        return HTTP_UNAVAILABLE_TEXT[reason as keyof typeof HTTP_UNAVAILABLE_TEXT];
    }
    return LOG_HISTORY_UNAVAILABLE_TEXT[reason as AppLogHistoryReason];
}

function HttpRequestsView({ metrics, isLoading, range }: HttpViewProps) {
    if (isLoading || !metrics) {
        return <AppLoader />;
    }
    if (!metrics.available) {
        const { reason } = metrics;

        return (
            <div className="flex flex-col gap-1 text-sm text-muted-foreground">
                <span>{httpUnavailableText(reason)}</span>
                {reason?.startsWith("access-log-") && (
                    <AppLink.Modules
                        to={ROUTE.systemSettings.traefik.configOptions.$route}
                        className="text-link"
                    >
                        Traefik Config Options
                    </AppLink.Modules>
                )}
                {(reason === "disabled" || reason === "apps-not-collected") && (
                    <AppLink.Modules
                        to={ROUTE.systemSettings.logging.configuration.$route}
                        className="text-link"
                    >
                        Logging settings
                    </AppLink.Modules>
                )}
            </div>
        );
    }

    return (
        <>
            {metrics.clamped && (
                <p className="text-xs text-muted-foreground">
                    The logs are kept for less than this range: the charts start where they do.
                </p>
            )}
            {metrics.totals && <HttpTotals totals={metrics.totals} />}
            <section className="flex flex-col gap-1">
                <h3 className="text-sm font-medium">Requests</h3>
                <RequestsChart
                    series={metrics.series}
                    range={range}
                />
            </section>
            <section className="flex flex-col gap-1">
                <h3 className="text-sm font-medium">Duration</h3>
                <DurationChart
                    series={metrics.series}
                    range={range}
                />
            </section>
            {metrics.byPath.length > 0 && (
                <section className="flex flex-col gap-1">
                    <h3 className="text-sm font-medium">Paths</h3>
                    <p className="text-xs text-muted-foreground">
                        The 20 most requested, numbers and ids counted as one: /users/:n, /orders/:id. The query string
                        is not logged.
                    </p>
                    <HttpPaths paths={metrics.byPath} />
                </section>
            )}
            {metrics.byReplica.length > 0 && (
                <section className="flex flex-col gap-1">
                    <h3 className="text-sm font-medium">Replicas</h3>
                    <HttpReplicas replicas={metrics.byReplica} />
                </section>
            )}
            <p className="text-xs text-muted-foreground">
                Counted from Traefik&apos;s access log: the requests that reach the app by its domains, from the client
                to its answer. Durations are close rather than exact.
            </p>
        </>
    );
}

function HttpTotals({ totals }: { totals: HttpMetricsCounts }) {
    const errorRate = totals.requests > 0 ? (totals.errors5xx / totals.requests) * 100 : 0;
    const items = [
        { label: "Requests", value: totals.requests.toLocaleString() },
        { label: "5xx", value: `${errorRate.toFixed(errorRate < 10 ? 2 : 1)}%` },
        { label: "p95", value: totals.p95 === null ? "-" : `${totals.p95.toFixed(1)} ms` },
        { label: "4xx", value: totals.errors4xx.toLocaleString() },
        { label: "Unreachable", value: totals.unreachable.toLocaleString() },
    ];

    return <TotalsGrid items={items} />;
}

function FunctionCallsView({ metrics, isLoading, range }: CallsViewProps) {
    if (isLoading || !metrics) {
        return <AppLoader />;
    }
    if (!metrics.available) {
        return (
            <div className="flex flex-col gap-1 text-sm text-muted-foreground">
                <span>
                    {metrics.reason
                        ? LOG_HISTORY_UNAVAILABLE_TEXT[metrics.reason]
                        : "The function's logs cannot be read."}
                </span>
                {(metrics.reason === "disabled" || metrics.reason === "apps-not-collected") && (
                    <AppLink.Modules
                        to={ROUTE.systemSettings.logging.configuration.$route}
                        className="text-link"
                    >
                        Logging settings
                    </AppLink.Modules>
                )}
            </div>
        );
    }

    return (
        <>
            {metrics.clamped && (
                <p className="text-xs text-muted-foreground">
                    The logs are kept for less than this range: the charts start where they do.
                </p>
            )}
            {metrics.totals && <FunctionTotals totals={metrics.totals} />}
            <section className="flex flex-col gap-1">
                <h3 className="text-sm font-medium">Calls</h3>
                <CallsChart
                    series={metrics.series}
                    range={range}
                />
            </section>
            <section className="flex flex-col gap-1">
                <h3 className="text-sm font-medium">Duration</h3>
                <DurationChart
                    series={metrics.series}
                    range={range}
                />
            </section>
            {Object.keys(metrics.byOutcome).length > 0 && (
                <section className="flex flex-wrap items-center gap-2 text-sm">
                    <span className="text-muted-foreground">By outcome:</span>
                    {Object.entries(metrics.byOutcome)
                        .sort(([, a], [, b]) => b - a)
                        .map(([outcome, calls]) => (
                            <Badge
                                key={outcome}
                                tone={outcome === "ok" ? "green" : "red"}
                                className="font-mono"
                            >
                                {outcome} {calls}
                            </Badge>
                        ))}
                </section>
            )}
            {metrics.byPath.length > 0 && (
                <section className="flex flex-col gap-1">
                    <h3 className="text-sm font-medium">Paths</h3>
                    <p className="text-xs text-muted-foreground">
                        The 20 most called. A path no handler serves - /.env, /.git/config - is a scanner trying the
                        function&apos;s domain; a handler answering 404 to it shows it under 4xx.
                    </p>
                    <MetricsPaths paths={metrics.byPath} />
                </section>
            )}
            <p className="text-xs text-muted-foreground">
                Counted from the line the runtime writes for every call: HTTP requests and scheduled calls through the
                function&apos;s server. Durations are the handler&apos;s, close rather than exact.
            </p>
        </>
    );
}

function FunctionTotals({ totals }: { totals: FunctionMetricsCounts }) {
    const failureRate = totals.calls > 0 ? (totals.failed / totals.calls) * 100 : 0;
    const items = [
        { label: "Calls", value: totals.calls.toLocaleString() },
        { label: "Failed", value: `${failureRate.toFixed(failureRate < 10 ? 2 : 1)}%` },
        { label: "p95", value: totals.p95 === null ? "-" : `${totals.p95.toFixed(1)} ms` },
        { label: "4xx", value: totals.errors4xx.toLocaleString() },
        { label: "5xx", value: totals.errors5xx.toLocaleString() },
    ];

    return <TotalsGrid items={items} />;
}

function TotalsGrid({ items }: { items: { label: string; value: string }[] }) {
    return (
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-5">
            {items.map(item => (
                <div
                    key={item.label}
                    className="flex flex-col gap-0.5 rounded-md border px-3 py-2"
                >
                    <span className="text-xs text-muted-foreground">{item.label}</span>
                    <span className="text-lg font-semibold tabular-nums">{item.value}</span>
                </div>
            ))}
        </div>
    );
}

interface HttpViewProps {
    metrics: AppLogs_GetHttpMetrics_Res["data"] | undefined;
    isLoading: boolean;
    range: FunctionMetricsRange;
}

interface CallsViewProps {
    metrics: AppLogs_GetFunctionMetrics_Res["data"] | undefined;
    isLoading: boolean;
    range: FunctionMetricsRange;
}
