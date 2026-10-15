import { useState } from "react";

import { cn } from "@/lib/utils";
import { ChevronRight, Undo2 } from "lucide-react";
import type { IgnoredFinding, VulnIgnoreRef } from "~/projects/api/services";

import { PopConfirm } from "@application/shared/components";

import { Button } from "@/components/ui";
import { Badge } from "@/components/ui/badge";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import { ignoreScopeLabel, severityLabel, severityTone } from "./severity";

interface Props {
    ignored: IgnoredFinding[];
    /** Remove on a row; none when the user may not. */
    onRemove?: (ignore: VulnIgnoreRef) => void;
}

/** What ignores set aside, folded: each finding with its ignore - where, why, who, until. */
export function IgnoredFindings({ ignored, onRemove }: Props) {
    const [open, setOpen] = useState(false);

    if (ignored.length === 0) {
        return null;
    }

    return (
        <Collapsible
            open={open}
            onOpenChange={setOpen}
        >
            <CollapsibleTrigger asChild>
                <button
                    type="button"
                    className="flex items-center gap-1.5 text-sm font-medium"
                >
                    <ChevronRight className={cn("size-4 transition-transform", open && "rotate-90")} />
                    Ignored ({ignored.length})
                </button>
            </CollapsibleTrigger>
            <CollapsibleContent className="pt-3">
                <div className="overflow-x-auto rounded-md border">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-24">Severity</TableHead>
                                <TableHead>Vulnerability</TableHead>
                                <TableHead>Package</TableHead>
                                <TableHead>Set aside for</TableHead>
                                <TableHead>Reason</TableHead>
                                <TableHead>Until</TableHead>
                                {onRemove && <TableHead className="w-24" />}
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {ignored.map(finding => (
                                <TableRow
                                    key={`${finding.imageId}|${finding.vulnId}|${finding.pkgName}|${finding.pkgPath}|${finding.installedVersion}`}
                                    data-testid="ignored-row"
                                >
                                    <TableCell>
                                        <Badge tone={severityTone(finding.severity)}>
                                            {severityLabel(finding.severity)}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className="font-mono text-xs">{finding.vulnId}</TableCell>
                                    <TableCell className="font-mono text-xs break-all">{finding.pkgName}</TableCell>
                                    <TableCell className="text-xs">
                                        {ignoreScopeLabel(finding.ignore)}
                                        {!finding.ignore.package && (
                                            <span className="text-muted-foreground"> · every package</span>
                                        )}
                                    </TableCell>
                                    <TableCell className="max-w-[24rem] text-xs">
                                        <span className="break-words">{finding.ignore.reason}</span>
                                        {finding.ignore.setBy && (
                                            <span className="text-muted-foreground"> · {finding.ignore.setBy}</span>
                                        )}
                                    </TableCell>
                                    <TableCell className="text-xs">
                                        {finding.ignore.expireAt
                                            ? finding.ignore.expireAt.toLocaleDateString()
                                            : "Never"}
                                    </TableCell>
                                    {onRemove && (
                                        <TableCell className="text-right">
                                            <PopConfirm
                                                title="Remove the ignore"
                                                description={`${finding.vulnId} counts again for every app the ignore covers.`}
                                                confirmText="Remove"
                                                variant="destructive"
                                                onConfirm={() => {
                                                    onRemove(finding.ignore);
                                                }}
                                            >
                                                <Button
                                                    type="button"
                                                    variant="ghost"
                                                    size="sm"
                                                    className="h-7 text-xs"
                                                >
                                                    <Undo2 className="size-3.5" />
                                                    Remove
                                                </Button>
                                            </PopConfirm>
                                        </TableCell>
                                    )}
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
}
