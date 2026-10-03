import { CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import type { FunctionMetricsRange, ResourceMetricsContainer, ResourceMetricsPoint } from "~/projects/api/services";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import { formatMetricsTime } from "./metrics-range";
import { formatBytes, formatCores } from "./resource-format";

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
function ResourceLineChart({ data, lines, format, width, replicas = false }: ChartProps) {
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
                    yAxisId="value"
                    tick={AXIS}
                    width={width}
                    tickFormatter={value => format(Number(value))}
                />
                {replicas && (
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
                    formatter={(value, _name, item) =>
                        typeof value === "number" && item.dataKey !== "replicas" ? format(value) : value
                    }
                />
                <Legend wrapperStyle={{ fontSize: 12 }} />
                {lines.map(line => (
                    <Line
                        key={line.key}
                        yAxisId="value"
                        dataKey={line.key}
                        name={line.name}
                        stroke={line.color}
                        strokeDasharray={line.dashed ? "4 4" : undefined}
                        dot={false}
                        strokeWidth={2}
                        isAnimationActive={false}
                    />
                ))}
                {replicas && (
                    <Line
                        yAxisId="replicas"
                        type="stepAfter"
                        dataKey="replicas"
                        name="Replicas"
                        stroke="var(--foreground)"
                        dot={false}
                        strokeWidth={1.5}
                        isAnimationActive={false}
                    />
                )}
            </LineChart>
        </ResponsiveContainer>
    );
}

/**
 * CPU in cores, the containers summed, and their limit when they have one; an autoscaled app's replicas, a line on
 * its own axis.
 */
export function CpuChart({ series, range }: SeriesProps) {
    const hasLimit = series.some(point => point.cpuLimit > 0);
    const hasReplicas = series.some(point => point.replicas !== null);
    const data = series.map(point => ({
        time: formatMetricsTime(point.time, range),
        cpu: point.cpu,
        limit: point.cpu === null || point.cpuLimit === 0 ? null : point.cpuLimit,
        replicas: point.replicas,
    }));

    return (
        <ResourceLineChart
            data={data}
            width={56}
            replicas={hasReplicas}
            format={value => value.toFixed(value < 1 ? 2 : 1)}
            lines={[
                { key: "cpu", name: "Used", color: "var(--chart-3)" },
                ...(hasLimit ? [{ key: "limit", name: "Limit", color: "var(--muted-foreground)", dashed: true }] : []),
            ]}
        />
    );
}

/** Memory, the containers' working sets summed, and their limit when they have one. */
export function MemoryChart({ series, range }: SeriesProps) {
    const hasLimit = series.some(point => point.memoryLimit > 0);
    const data = series.map(point => ({
        time: formatMetricsTime(point.time, range),
        memory: point.memory,
        limit: point.memory === null || point.memoryLimit === 0 ? null : point.memoryLimit,
    }));

    return (
        <ResourceLineChart
            data={data}
            width={72}
            format={value => formatBytes(value)}
            lines={[
                { key: "memory", name: "Used", color: "var(--chart-4)" },
                ...(hasLimit ? [{ key: "limit", name: "Limit", color: "var(--muted-foreground)", dashed: true }] : []),
            ]}
        />
    );
}

/** Network in and out, in bytes a second, the containers summed. */
export function NetworkChart({ series, range }: SeriesProps) {
    const data = series.map(point => ({
        time: formatMetricsTime(point.time, range),
        rx: point.cpu === null ? null : point.netRx,
        tx: point.cpu === null ? null : point.netTx,
    }));

    return (
        <ResourceLineChart
            data={data}
            width={72}
            format={value => `${formatBytes(value)}/s`}
            lines={[
                { key: "rx", name: "In", color: "var(--chart-2)" },
                { key: "tx", name: "Out", color: "var(--chart-5)" },
            ]}
        />
    );
}

/** The app's containers over the range, the last seen first. */
export function ResourceContainers({ containers }: { containers: ResourceMetricsContainer[] }) {
    return (
        <div className="overflow-x-auto rounded-md border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Container</TableHead>
                        <TableHead className="text-right">CPU</TableHead>
                        <TableHead className="text-right">CPU peak</TableHead>
                        <TableHead className="text-right">Memory peak</TableHead>
                        <TableHead className="text-right">OOM kills</TableHead>
                        <TableHead className="text-right">Last seen</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {containers.map(item => (
                        <TableRow key={item.container}>
                            <TableCell className="font-mono text-xs">{item.container}</TableCell>
                            <TableCell className="text-right tabular-nums">{formatCores(item.cpu)}</TableCell>
                            <TableCell className="text-right tabular-nums">{formatCores(item.cpuPeak)}</TableCell>
                            <TableCell className="text-right tabular-nums">
                                {formatBytes(item.memory)}
                                {item.memoryLimit > 0 && (
                                    <span className="text-muted-foreground"> / {formatBytes(item.memoryLimit)}</span>
                                )}
                            </TableCell>
                            <TableCell className="text-right tabular-nums">{item.oomKills}</TableCell>
                            <TableCell className="text-right text-xs text-muted-foreground">
                                {item.lastSeen ? new Date(item.lastSeen).toLocaleString() : "-"}
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}

interface SeriesProps {
    series: ResourceMetricsPoint[];
    range: FunctionMetricsRange;
}

interface ChartProps {
    data: Record<string, string | number | null>[];
    lines: { key: string; name: string; color: string; dashed?: boolean }[];
    format: (value: number) => string;
    width: number;
    /** Draws the data's replicas, on an axis of their own. */
    replicas?: boolean;
}
