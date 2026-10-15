import { useMemo, useState } from "react";

import { inlineLink } from "@lib/styles";
import { EyeOff } from "lucide-react";
import type { VulnFinding } from "~/projects/api/services";

import { Button, Checkbox, Input } from "@/components/ui";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import { severityLabel, severityTone } from "./severity";

interface Props {
    findings: VulnFinding[];
    /** Ignore… on a row; none when the user may not set one aside. */
    onIgnore?: (finding: VulnFinding) => void;
}

/**
 * The findings no ignore sets aside, the worst first. Those with a fixed
 * version are shown unless asked: they are what a rebuild takes away.
 */
export function FindingsTable({ findings, onIgnore }: Props) {
    const [fixableOnly, setFixableOnly] = useState(true);
    const [search, setSearch] = useState("");

    const shown = useMemo(() => {
        const term = search.trim().toLowerCase();
        return findings.filter(
            finding =>
                (!fixableOnly || finding.fixedVersion !== "") &&
                (term === "" ||
                    finding.vulnId.toLowerCase().includes(term) ||
                    finding.pkgName.toLowerCase().includes(term) ||
                    finding.title.toLowerCase().includes(term)),
        );
    }, [findings, fixableOnly, search]);

    return (
        <div className="flex flex-col gap-3">
            <div className="flex flex-wrap items-center gap-3">
                <Input
                    value={search}
                    onChange={event => {
                        setSearch(event.target.value);
                    }}
                    placeholder="Filter by vulnerability, package or title"
                    className="h-8 w-full max-w-[320px]"
                    aria-label="Filter findings"
                />
                <label
                    htmlFor="findings-fixable-only"
                    className="flex cursor-pointer select-none items-center gap-2 text-sm"
                >
                    <Checkbox
                        id="findings-fixable-only"
                        checked={fixableOnly}
                        onCheckedChange={checked => {
                            setFixableOnly(checked === true);
                        }}
                    />
                    Fixable only
                </label>
                <span className="text-xs text-muted-foreground">
                    {shown.length} of {findings.length}
                </span>
            </div>

            {shown.length === 0 ? (
                <div className="rounded-md border px-4 py-6 text-center text-sm text-muted-foreground">
                    {findings.length === 0
                        ? "No vulnerability found."
                        : fixableOnly
                          ? "Nothing here has a fixed version. Clear Fixable only to see the rest."
                          : "Nothing matches the filter."}
                </div>
            ) : (
                <div className="overflow-x-auto rounded-md border">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-24">Severity</TableHead>
                                <TableHead>Vulnerability</TableHead>
                                <TableHead>Package</TableHead>
                                <TableHead>Installed</TableHead>
                                <TableHead>Fixed in</TableHead>
                                {onIgnore && <TableHead className="w-24" />}
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {shown.map(finding => (
                                <TableRow
                                    key={`${finding.imageId}|${finding.vulnId}|${finding.pkgName}|${finding.pkgPath}|${finding.installedVersion}`}
                                    data-testid="finding-row"
                                >
                                    <TableCell>
                                        <Badge tone={severityTone(finding.severity)}>
                                            {severityLabel(finding.severity)}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className="max-w-[28rem]">
                                        <div className="flex flex-col gap-0.5">
                                            {finding.url ? (
                                                <a
                                                    href={finding.url}
                                                    target="_blank"
                                                    rel="noreferrer"
                                                    className={`font-mono text-xs ${inlineLink}`}
                                                >
                                                    {finding.vulnId}
                                                </a>
                                            ) : (
                                                <span className="font-mono text-xs">{finding.vulnId}</span>
                                            )}
                                            {finding.title && (
                                                <span
                                                    className="truncate text-xs text-muted-foreground"
                                                    title={finding.title}
                                                >
                                                    {finding.title}
                                                </span>
                                            )}
                                        </div>
                                    </TableCell>
                                    <TableCell>
                                        <div className="flex flex-col gap-0.5">
                                            <span className="font-mono text-xs break-all">{finding.pkgName}</span>
                                            <span
                                                className="text-xs text-muted-foreground break-all"
                                                title={finding.pkgPath}
                                            >
                                                {finding.pkgType}
                                                {finding.pkgPath && ` · ${finding.pkgPath}`}
                                            </span>
                                        </div>
                                    </TableCell>
                                    <TableCell className="font-mono text-xs">{finding.installedVersion}</TableCell>
                                    <TableCell className="font-mono text-xs">
                                        {finding.fixedVersion || (
                                            <span className="text-muted-foreground">No fix yet</span>
                                        )}
                                    </TableCell>
                                    {onIgnore && (
                                        <TableCell className="text-right">
                                            <Button
                                                type="button"
                                                variant="ghost"
                                                size="sm"
                                                className="h-7 text-xs"
                                                onClick={() => {
                                                    onIgnore(finding);
                                                }}
                                            >
                                                <EyeOff className="size-3.5" />
                                                Ignore…
                                            </Button>
                                        </TableCell>
                                    )}
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </div>
            )}
        </div>
    );
}
