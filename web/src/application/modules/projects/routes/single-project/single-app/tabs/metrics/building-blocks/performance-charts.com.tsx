import { Bar, CartesianGrid, ComposedChart, Legend, Line, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import type { FunctionMetricsRange, PerformancePoint } from "~/projects/api/services";

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

/**
 * Requests or calls as OBI saw them, those that failed apart, one bar per step; an autoscaled app's replicas, a
 * line on its own axis.
 */
export function PerformanceCountChart({ series, range, label }: Props) {
    const data = series.map(point => ({
        time: formatMetricsTime(point.time, range),
        ok: point.requests - point.errors,
        errors: point.errors,
        replicas: point.replicas,
    }));
    const hasReplicas = series.some(point => point.replicas !== null);

    return (
        <ResponsiveContainer
            width="100%"
            height={CHART_HEIGHT}
        >
            <ComposedChart data={data}>
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
                    yAxisId="count"
                    tick={AXIS}
                    allowDecimals={false}
                    width={40}
                />
                {hasReplicas && (
                    <YAxis
                        yAxisId="replicas"
                        orientation="right"
                        tick={AXIS}
                        allowDecimals={false}
                        domain={[0, "dataMax + 1"]}
                        width={32}
                    />
                )}
                <Tooltip
                    {...TOOLTIP}
                    cursor={{ fill: "var(--muted)" }}
                />
                <Legend wrapperStyle={{ fontSize: 12 }} />
                <Bar
                    yAxisId="count"
                    dataKey="ok"
                    name={label}
                    stackId="count"
                    fill="var(--chart-2)"
                />
                <Bar
                    yAxisId="count"
                    dataKey="errors"
                    name="Failed"
                    stackId="count"
                    fill="var(--destructive)"
                />
                {hasReplicas && (
                    <Line
                        yAxisId="replicas"
                        type="stepAfter"
                        dataKey="replicas"
                        name="Replicas"
                        stroke="var(--chart-3)"
                        dot={false}
                        strokeWidth={2}
                        isAnimationActive={false}
                    />
                )}
            </ComposedChart>
        </ResponsiveContainer>
    );
}

interface Props {
    series: PerformancePoint[];
    range: FunctionMetricsRange;
    /** What a bar counts that did not fail: Requests, Calls. */
    label: string;
}
