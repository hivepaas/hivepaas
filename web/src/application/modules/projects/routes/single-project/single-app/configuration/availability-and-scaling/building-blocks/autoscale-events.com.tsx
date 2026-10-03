import type { AppAutoscaleEvent } from "~/projects/api/services";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

/** The latest scalings: a function's from its calls, any other app's from its requests and CPU. */
export function AutoscaleEvents({ events, isFunction }: Props) {
    return (
        <div className="max-w-[800px] overflow-x-auto rounded-md border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Time</TableHead>
                        <TableHead className="text-right">Replicas</TableHead>
                        <TableHead className="text-right">In flight</TableHead>
                        <TableHead className="text-right">{isFunction ? "Turned away" : "CPU"}</TableHead>
                        <TableHead>Why</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {events.map(event => (
                        <TableRow key={`${event.time}-${event.from}-${event.to}`}>
                            <TableCell className="text-xs text-muted-foreground">
                                {new Date(event.time).toLocaleString()}
                            </TableCell>
                            <TableCell className="text-right tabular-nums">
                                {event.from} → {event.to}
                            </TableCell>
                            <TableCell className="text-right tabular-nums">
                                {isFunction || event.requests > 0 || event.inFlight > 0
                                    ? event.inFlight.toFixed(1)
                                    : "–"}
                            </TableCell>
                            <TableCell className="text-right tabular-nums">
                                {isFunction ? event.throttled : event.cpu > 0 ? `${event.cpu.toFixed(0)} %` : "–"}
                            </TableCell>
                            <TableCell className="text-xs">{event.reason}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}

interface Props {
    events: AppAutoscaleEvent[];
    isFunction: boolean;
}
