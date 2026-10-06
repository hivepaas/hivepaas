import type {
    AppLogHistoryReason,
    AppLogs_GetDependencyMetrics_Res,
    AppLogs_GetRouteMetrics_Res,
    AppPerformanceMetricsHead,
    FunctionMetricsRange,
} from "~/projects/api/services";

import { AppLink, AppLoader } from "@application/shared/components";
import { ROUTE, obiPreflightReasonText } from "@application/shared/constants";

import { LOG_HISTORY_UNAVAILABLE_TEXT } from "../../logs/building-blocks";

import { DurationChart } from "./metrics-charts.com";
import { PERFORMANCE_UNAVAILABLE_TEXT } from "./metrics-unavailable.constants";
import { PerformanceCountChart } from "./performance-charts.com";
import { callKindLabel, formatMs, formatShare } from "./performance-format";
import { PeersTable, RoutesTable } from "./performance-tables.com";
import { TotalsGrid } from "./totals-grid.com";

/** Why an app's routes and calls cannot be shown, and where to change it. */
function PerformanceUnavailable({ head, projectID, env, appID }: AppRefProps & { head: AppPerformanceMetricsHead }) {
    const { reason } = head;
    const text = !reason
        ? "The app's routes and calls cannot be read."
        : reason in PERFORMANCE_UNAVAILABLE_TEXT
          ? PERFORMANCE_UNAVAILABLE_TEXT[reason as keyof typeof PERFORMANCE_UNAVAILABLE_TEXT]
          : LOG_HISTORY_UNAVAILABLE_TEXT[reason as AppLogHistoryReason];
    const toSystem = reason === "performance-disabled" || reason === "node-disabled" || reason === "node-unsupported";

    return (
        <div className="flex flex-col gap-1 text-sm text-muted-foreground">
            <span>{text}</span>
            {reason === "node-unsupported" && head.preflightReasons.length > 0 && (
                <ul className="list-disc pl-5">
                    {head.preflightReasons.map(item => (
                        <li key={item}>{obiPreflightReasonText(item)}</li>
                    ))}
                </ul>
            )}
            {reason === "app-disabled" && (
                <AppLink.Basic
                    to={ROUTE.projects.single.apps.single.configuration.featureSettings.$route(projectID, env, appID)}
                    className="text-link"
                >
                    Feature Settings
                </AppLink.Basic>
            )}
            {toSystem && (
                <AppLink.Modules
                    to={ROUTE.systemSettings.logging.performance.$route}
                    className="text-link"
                >
                    Routes and Calls settings
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

/** What the numbers cover: the logs' range, and the nodes that measure the app. */
function PerformanceNotes({ head }: { head: AppPerformanceMetricsHead }) {
    return (
        <>
            {head.clamped && (
                <p className="text-xs text-muted-foreground">
                    The logs are kept for less than this range: the charts start where they do.
                </p>
            )}
            {head.nodes === 0 && (
                <p className="text-xs text-muted-foreground">
                    The app runs on no node now: these numbers are from before.
                </p>
            )}
            {head.nodes > 0 && head.nodesCovered < head.nodes && (
                <p className="text-xs text-muted-foreground">
                    {head.nodesCovered} of the {head.nodes} nodes this app runs on measure it: what it serves and calls
                    on the others is not counted.
                </p>
            )}
        </>
    );
}

/** What an app served, by route, as OBI saw it in its containers. */
export function RoutesView({ metrics, isLoading, range, projectID, env, appID }: RoutesViewProps) {
    if (isLoading || !metrics) {
        return <AppLoader />;
    }
    if (!metrics.available) {
        return (
            <PerformanceUnavailable
                head={metrics}
                projectID={projectID}
                env={env}
                appID={appID}
            />
        );
    }
    const { totals } = metrics;

    return (
        <>
            <PerformanceNotes head={metrics} />
            {totals && (
                <TotalsGrid
                    items={[
                        { label: "Requests", value: totals.requests.toLocaleString() },
                        { label: "Failed", value: formatShare(totals.errors, totals.requests) },
                        { label: "p50", value: formatMs(totals.p50) },
                        { label: "p95", value: formatMs(totals.p95) },
                        { label: "p99", value: formatMs(totals.p99) },
                    ]}
                />
            )}
            <section className="flex flex-col gap-1">
                <h3 className="text-sm font-medium">Requests</h3>
                <PerformanceCountChart
                    series={metrics.series}
                    range={range}
                    label="Requests"
                />
            </section>
            <section className="flex flex-col gap-1">
                <h3 className="text-sm font-medium">Duration</h3>
                <DurationChart
                    series={metrics.series}
                    range={range}
                />
            </section>
            {metrics.routes.length > 0 && (
                <section className="flex flex-col gap-1">
                    <h3 className="text-sm font-medium">Routes</h3>
                    <p className="text-xs text-muted-foreground">
                        The 100 most requested, by route as the app&apos;s framework names it - /users/{"{id}"} - or by
                        path when it names none.
                    </p>
                    <RoutesTable routes={metrics.routes} />
                </section>
            )}
            <p className="text-xs text-muted-foreground">
                Measured inside the app&apos;s containers by OBI (eBPF): every request it answered, from Traefik, from
                another app of the project, or from its own health check - a function&apos;s is left out. A request
                fails on a 5xx or an error. Durations are the app&apos;s own, close rather than exact.
            </p>
        </>
    );
}

/** What an app called - other apps, databases, outside hosts - as OBI saw it in its containers. */
export function DependenciesView({ metrics, isLoading, range, projectID, env, appID }: DependenciesViewProps) {
    if (isLoading || !metrics) {
        return <AppLoader />;
    }
    if (!metrics.available) {
        return (
            <PerformanceUnavailable
                head={metrics}
                projectID={projectID}
                env={env}
                appID={appID}
            />
        );
    }

    return (
        <>
            <PerformanceNotes head={metrics} />
            {metrics.kinds.length === 0 && (
                <p className="text-sm text-muted-foreground">The app called nothing OBI measures in this range.</p>
            )}
            {metrics.kinds.map(kind => (
                <section
                    key={kind.kind}
                    className="flex flex-col gap-2"
                >
                    <h3 className="text-sm font-medium">{callKindLabel(kind.kind)} calls</h3>
                    <TotalsGrid
                        items={[
                            { label: "Calls", value: kind.totals.requests.toLocaleString() },
                            { label: "Failed", value: formatShare(kind.totals.errors, kind.totals.requests) },
                            { label: "p50", value: formatMs(kind.totals.p50) },
                            { label: "p95", value: formatMs(kind.totals.p95) },
                            { label: "p99", value: formatMs(kind.totals.p99) },
                        ]}
                    />
                    <PerformanceCountChart
                        series={kind.series}
                        range={range}
                        label="Calls"
                    />
                </section>
            ))}
            {metrics.peers.length > 0 && (
                <section className="flex flex-col gap-1">
                    <h3 className="text-sm font-medium">Peers</h3>
                    <PeersTable
                        peers={metrics.peers}
                        projectID={projectID}
                        env={env}
                    />
                </section>
            )}
            <p className="text-xs text-muted-foreground">
                Measured inside the app&apos;s containers by OBI (eBPF): every call it made, by the name it called - an
                app of this environment is shown as itself - or, for a database, by its system and database. No query
                text is kept. Durations are close rather than exact.
            </p>
        </>
    );
}

interface AppRefProps {
    projectID: string;
    env: string;
    appID: string;
}

interface RoutesViewProps extends AppRefProps {
    metrics: AppLogs_GetRouteMetrics_Res["data"] | undefined;
    isLoading: boolean;
    range: FunctionMetricsRange;
}

interface DependenciesViewProps extends AppRefProps {
    metrics: AppLogs_GetDependencyMetrics_Res["data"] | undefined;
    isLoading: boolean;
    range: FunctionMetricsRange;
}
