import {
    Bar,
    BarChart,
    CartesianGrid,
    Legend,
    Line,
    LineChart,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis,
} from "recharts";
import type { FunctionMetricsPoint, FunctionMetricsRange } from "~/projects/api/services";

import { formatMetricsTime } from "./metrics-range";

const AXIS = { fontSize: 11, fill: "var(--muted-foreground)" };
const TOOLTIP = {
    contentStyle: {
        background: "var(--popover)",
        border: "1px solid var(--border)",
        borderRadius: 6,
        fontSize: 12,
        color: "var(--popover-foreground)",
    },
    labelStyle: { color: "var(--muted-foreground)" },
};
const CHART_HEIGHT = 220;

/** Calls and failed calls, one bar per step. */
export function CallsChart({ series, range }: Props) {
    const data = series.map(point => ({
        time: formatMetricsTime(point.time, range),
        ok: point.calls - point.failed,
        failed: point.failed,
    }));

    return (
        <ResponsiveContainer
            width="100%"
            height={CHART_HEIGHT}
        >
            <BarChart data={data}>
                <CartesianGrid
                    stroke="var(--border)"
                    vertical={false}
                />
                <XAxis
                    dataKey="time"
                    tick={AXIS}
                    minTickGap={24}
                />
                <YAxis
                    tick={AXIS}
                    allowDecimals={false}
                    width={40}
                />
                <Tooltip
                    {...TOOLTIP}
                    cursor={{ fill: "var(--muted)" }}
                />
                <Legend wrapperStyle={{ fontSize: 12 }} />
                <Bar
                    dataKey="ok"
                    name="Succeeded"
                    stackId="calls"
                    fill="var(--chart-2)"
                />
                <Bar
                    dataKey="failed"
                    name="Failed"
                    stackId="calls"
                    fill="var(--destructive)"
                />
            </BarChart>
        </ResponsiveContainer>
    );
}

/** p50, p95 and p99 of a duration; a step without a call or a request is a gap. */
export function DurationChart({ series, range }: DurationProps) {
    const data = series.map(point => ({
        time: formatMetricsTime(point.time, range),
        p50: point.p50,
        p95: point.p95,
        p99: point.p99,
    }));

    return (
        <ResponsiveContainer
            width="100%"
            height={CHART_HEIGHT}
        >
            <LineChart data={data}>
                <CartesianGrid
                    stroke="var(--border)"
                    vertical={false}
                />
                <XAxis
                    dataKey="time"
                    tick={AXIS}
                    minTickGap={24}
                />
                <YAxis
                    tick={AXIS}
                    width={48}
                    unit=" ms"
                />
                <Tooltip
                    {...TOOLTIP}
                    formatter={value => (typeof value === "number" ? `${value.toFixed(1)} ms` : value)}
                />
                <Legend wrapperStyle={{ fontSize: 12 }} />
                {(["p50", "p95", "p99"] as const).map((key, index) => (
                    <Line
                        key={key}
                        dataKey={key}
                        name={key}
                        stroke={`var(--chart-${index + 3})`}
                        dot={false}
                        strokeWidth={2}
                        isAnimationActive={false}
                    />
                ))}
            </LineChart>
        </ResponsiveContainer>
    );
}

interface Props {
    series: FunctionMetricsPoint[];
    range: FunctionMetricsRange;
}

/** A point with durations: a function's calls, or an app's requests. */
interface DurationPoint {
    time: string;
    p50: number | null;
    p95: number | null;
    p99: number | null;
}

interface DurationProps {
    series: DurationPoint[];
    range: FunctionMetricsRange;
}
