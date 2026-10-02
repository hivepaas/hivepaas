import { useState } from "react";

import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { RefreshCw } from "lucide-react";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import type { FunctionMetricsCounts, FunctionMetricsRange } from "~/projects/api/services";
import { AppLogsQueries } from "~/projects/data";

import { AppLink, AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

import { LOG_HISTORY_UNAVAILABLE_TEXT } from "../../logs/building-blocks";
import {
    CallsChart,
    DurationChart,
    METRICS_RANGES,
    MetricsPaths,
    storeMetricsRange,
    storedMetricsRange,
} from "../building-blocks";

/**
 * A function's calls over a range ending now - how many, how many failed, how
 * long the handler ran - counted from the invocation line its runtime writes
 * for every call into its logs.
 */
export function AppMetricsRoute() {
    const { id: projectID, env, appId: appID } = useParams<{ id: string; env: string; appId: string }>();
    invariant(projectID, "projectID must be defined");
    invariant(env, "env must be defined");
    invariant(appID, "appID must be defined");

    const [range, setRange] = useState<FunctionMetricsRange>(storedMetricsRange);
    const query = AppLogsQueries.useGetFunctionMetrics({ projectID, env, appID, range });
    const metrics = query.data?.data;

    function chooseRange(next: FunctionMetricsRange) {
        setRange(next);
        storeMetricsRange(next);
    }

    return (
        <div className={cn(listBox, "flex flex-col gap-4")}>
            <div className="flex flex-wrap items-center justify-between gap-2">
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
                <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    isLoading={query.isFetching}
                    onClick={() => {
                        void query.refetch();
                    }}
                >
                    <RefreshCw /> Refresh
                </Button>
            </div>

            {query.isLoading || !metrics ? (
                <AppLoader />
            ) : !metrics.available ? (
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
            ) : (
                <>
                    {metrics.clamped && (
                        <p className="text-xs text-muted-foreground">
                            The logs are kept for less than this range: the charts start where they do.
                        </p>
                    )}
                    {metrics.totals && <MetricsTotals totals={metrics.totals} />}
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
                                The 20 most called. A path no handler serves - /.env, /.git/config - is a scanner trying
                                the function&apos;s domain; a handler answering 404 to it shows it under 4xx.
                            </p>
                            <MetricsPaths paths={metrics.byPath} />
                        </section>
                    )}
                    <p className="text-xs text-muted-foreground">
                        Counted from the line the runtime writes for every call: HTTP requests and scheduled calls
                        through the function&apos;s server. Durations are the handler&apos;s, close rather than exact.
                    </p>
                </>
            )}
        </div>
    );
}

function MetricsTotals({ totals }: { totals: FunctionMetricsCounts }) {
    const failureRate = totals.calls > 0 ? (totals.failed / totals.calls) * 100 : 0;
    const items = [
        { label: "Calls", value: totals.calls.toLocaleString() },
        { label: "Failed", value: `${failureRate.toFixed(failureRate < 10 ? 2 : 1)}%` },
        { label: "p95", value: totals.p95 === null ? "-" : `${totals.p95.toFixed(1)} ms` },
        { label: "4xx", value: totals.errors4xx.toLocaleString() },
        { label: "5xx", value: totals.errors5xx.toLocaleString() },
    ];

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
