import { Bar, CartesianGrid, ComposedChart, Legend, Line, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import type { FunctionMetricsRange, HttpMetricsPoint } from "~/projects/api/services";

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
 * Requests by what the client got - 2xx and 3xx, 4xx, 5xx - one bar per step; an autoscaled app's replicas, a
 * line on its own axis.
 */
export function RequestsChart({ series, range }: Props) {
    const data = series.map(point => ({
        time: formatMetricsTime(point.time, range),
        ok: point.requests - point.errors4xx - point.errors5xx,
        errors4xx: point.errors4xx,
        errors5xx: point.errors5xx,
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
                    yAxisId="requests"
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
                    yAxisId="requests"
                    dataKey="ok"
                    name="2xx, 3xx"
                    stackId="requests"
                    fill="var(--chart-2)"
                />
                <Bar
                    yAxisId="requests"
                    dataKey="errors4xx"
                    name="4xx"
                    stackId="requests"
                    fill="var(--chart-4)"
                />
                <Bar
                    yAxisId="requests"
                    dataKey="errors5xx"
                    name="5xx"
                    stackId="requests"
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
    series: HttpMetricsPoint[];
    range: FunctionMetricsRange;
}
