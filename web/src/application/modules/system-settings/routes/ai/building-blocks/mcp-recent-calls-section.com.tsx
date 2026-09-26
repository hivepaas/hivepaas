import { Badge } from "@components/ui/badge";
import { Link } from "react-router";
import { AuditLogsQueries } from "~/operations/data";
import { AuditLogResult, AuditLogType } from "~/operations/domain";
import { SectionHeader } from "~/system-settings/module-shared";

import { ROUTE } from "@application/shared/constants";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui";

const RECENT_CALLS = 20;

/**
 * What assistants have been asking, from the audit log: every tool call is an
 * entry, a refused one included. So a person can see what their assistant did,
 * and what somebody else's did with a key.
 */
export function McpRecentCallsSection() {
    const { data, isLoading } = AuditLogsQueries.useFindManyPaginated({
        scope: { type: "global" },
        pagination: { page: 1, size: RECENT_CALLS },
        type: [AuditLogType.McpToolCall],
    });
    const calls = data?.data ?? [];

    return (
        <>
            <SectionHeader>Recent calls</SectionHeader>
            <div className="flex flex-col gap-2 px-3">
                {calls.length === 0 ? (
                    <span className="text-sm text-muted-foreground">
                        {isLoading ? "Loading…" : "No assistant has called a tool yet."}
                    </span>
                ) : (
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Time</TableHead>
                                <TableHead>User</TableHead>
                                <TableHead>Tool</TableHead>
                                <TableHead>Result</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {calls.map(call => (
                                <TableRow key={call.id}>
                                    <TableCell className="whitespace-nowrap text-muted-foreground">
                                        {call.createdAt.toLocaleString()}
                                    </TableCell>
                                    <TableCell>{call.actor?.name ?? call.actor?.loggedName ?? "-"}</TableCell>
                                    <TableCell className="font-mono text-xs">
                                        {call.resource?.loggedName ?? call.resource?.name ?? "-"}
                                    </TableCell>
                                    <TableCell>
                                        {call.result === AuditLogResult.Denied ? (
                                            <Badge variant="destructive">Refused</Badge>
                                        ) : (
                                            <Badge variant="secondary">Answered</Badge>
                                        )}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                )}
                <Link
                    to={ROUTE.operations.auditLogs.$route}
                    className="self-start text-xs text-muted-foreground underline underline-offset-2"
                >
                    Every call, with what it asked, is in the audit log
                </Link>
            </div>
        </>
    );
}
