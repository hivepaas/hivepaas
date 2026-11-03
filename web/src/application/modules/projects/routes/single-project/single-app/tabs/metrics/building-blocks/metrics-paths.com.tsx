import type { FunctionMetricsPath } from "~/projects/api/services";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

/**
 * The most called methods and paths: a function's own, and the scanners' that
 * try every domain they learn of - /.env, /.git/config, /wp-login.php.
 */
export function MetricsPaths({ paths }: { paths: FunctionMetricsPath[] }) {
    return (
        <div className="overflow-x-auto rounded-md border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead className="w-20">Method</TableHead>
                        <TableHead>Path</TableHead>
                        <TableHead className="text-right">Calls</TableHead>
                        <TableHead className="text-right">Failed</TableHead>
                        <TableHead className="text-right">4xx</TableHead>
                        <TableHead className="text-right">5xx</TableHead>
                        <TableHead className="text-right">p95</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {paths.map(item => (
                        <TableRow key={`${item.method} ${item.path}`}>
                            <TableCell className="font-mono text-xs">{item.method}</TableCell>
                            <TableCell
                                className="max-w-[28rem] truncate font-mono text-xs"
                                title={item.path}
                            >
                                {item.path}
                            </TableCell>
                            <TableCell className="text-right tabular-nums">{item.calls.toLocaleString()}</TableCell>
                            <TableCell className="text-right tabular-nums">{item.failed.toLocaleString()}</TableCell>
                            <TableCell className="text-right tabular-nums">{item.errors4xx.toLocaleString()}</TableCell>
                            <TableCell className="text-right tabular-nums">{item.errors5xx.toLocaleString()}</TableCell>
                            <TableCell className="text-right tabular-nums">
                                {item.p95 === null ? "-" : `${item.p95.toFixed(1)} ms`}
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}
