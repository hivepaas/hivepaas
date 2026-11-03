import type { HttpMetricsPath, HttpMetricsReplica } from "~/projects/api/services";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

function ms(value: number | null) {
    return value === null ? "-" : `${value.toFixed(1)} ms`;
}

/** The most requested methods and paths, numbers and ids counted as one. */
export function HttpPaths({ paths }: { paths: HttpMetricsPath[] }) {
    return (
        <div className="overflow-x-auto rounded-md border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead className="w-20">Method</TableHead>
                        <TableHead>Path</TableHead>
                        <TableHead className="text-right">Requests</TableHead>
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
                            <TableCell className="text-right tabular-nums">{item.requests.toLocaleString()}</TableCell>
                            <TableCell className="text-right tabular-nums">{item.errors4xx.toLocaleString()}</TableCell>
                            <TableCell className="text-right tabular-nums">{item.errors5xx.toLocaleString()}</TableCell>
                            <TableCell className="text-right tabular-nums">{ms(item.p95)}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}

/** The requests each replica answered, by its address in the project's network. */
export function HttpReplicas({ replicas }: { replicas: HttpMetricsReplica[] }) {
    return (
        <div className="overflow-x-auto rounded-md border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Replica</TableHead>
                        <TableHead className="text-right">Requests</TableHead>
                        <TableHead className="text-right">5xx</TableHead>
                        <TableHead className="text-right">p95</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {replicas.map(item => (
                        <TableRow key={item.address}>
                            <TableCell className="font-mono text-xs">
                                {item.address || <span className="text-muted-foreground">No replica answered</span>}
                            </TableCell>
                            <TableCell className="text-right tabular-nums">{item.requests.toLocaleString()}</TableCell>
                            <TableCell className="text-right tabular-nums">{item.errors5xx.toLocaleString()}</TableCell>
                            <TableCell className="text-right tabular-nums">{ms(item.p95)}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}
