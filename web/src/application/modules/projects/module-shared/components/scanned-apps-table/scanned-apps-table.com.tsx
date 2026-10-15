import { cn } from "@lib/utils";
import { AlertTriangle } from "lucide-react";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

interface VulnCount {
    total: number;
    fixable: number;
}

/** An app and what its images carry: what the table shows of it. */
export interface ScannedAppRow {
    app: { id: string; name: string };
    project: { id: string; name: string };
    projectEnv: { id: string; key: string; name: string };
    scannedAt: Date;
    error: string;
    summary: { critical: VulnCount; high: VulnCount; medium: VulnCount; ignored: number };
    images: { imageId: string; imageRef: string; osFamily: string; osName: string; osEosl: boolean; status: string }[];
}

function Count({ count, tone }: { count: VulnCount; tone: string }) {
    return (
        <span className="tabular-nums">
            <span className={cn(count.fixable > 0 && tone, count.fixable > 0 && "font-semibold")}>{count.fixable}</span>
            <span className="text-muted-foreground"> / {count.total}</span>
        </span>
    );
}

function ImagesCell({ item }: { item: ScannedAppRow }) {
    return (
        <div className="flex flex-col gap-0.5">
            {item.images.map(image => (
                <span
                    key={image.imageId}
                    className="font-mono text-xs break-all"
                    title={image.imageId}
                >
                    {image.imageRef || image.imageId.slice(0, 19)}
                    {image.osEosl && (
                        <span className="ml-1 font-sans text-amber-700 dark:text-amber-400">
                            · {image.osFamily} {image.osName} unsupported
                        </span>
                    )}
                    {image.status === "failed" && (
                        <span className="ml-1 font-sans text-destructive">· scan failed</span>
                    )}
                </span>
            ))}
            {item.error && (
                <span className="flex items-start gap-1 text-xs text-destructive">
                    <AlertTriangle className="mt-0.5 size-3 shrink-0" />
                    {item.error}
                </span>
            )}
        </div>
    );
}

/**
 * Scanned apps and what their images carry, in the order given - the worst
 * first. Counts are after ignores, fixable / all; each app links to its
 * Security tab. withProject names each app's project as well as its env.
 */
export function ScannedAppsTable({ items, withProject = false }: { items: ScannedAppRow[]; withProject?: boolean }) {
    return (
        <div className="overflow-x-auto rounded-md border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>App</TableHead>
                        <TableHead>Images</TableHead>
                        <TableHead className="text-right">Critical</TableHead>
                        <TableHead className="text-right">High</TableHead>
                        <TableHead className="text-right">Medium</TableHead>
                        <TableHead className="text-right">Ignored</TableHead>
                        <TableHead>Scanned</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {items.map(item => (
                        <TableRow
                            key={item.app.id}
                            data-testid="image-scan-app-row"
                        >
                            <TableCell>
                                <div className="flex flex-col gap-0.5">
                                    <AppLink.Basic
                                        to={ROUTE.projects.single.apps.single.security.$route(
                                            item.project.id,
                                            item.projectEnv.key,
                                            item.app.id,
                                        )}
                                        className="font-medium underline-offset-4 hover:underline"
                                    >
                                        {item.app.name}
                                    </AppLink.Basic>
                                    <span className="text-xs text-muted-foreground">
                                        {withProject
                                            ? `${item.project.name} / ${item.projectEnv.name}`
                                            : item.projectEnv.name}
                                    </span>
                                </div>
                            </TableCell>
                            <TableCell className="max-w-[26rem]">
                                <ImagesCell item={item} />
                            </TableCell>
                            <TableCell className="text-right">
                                <Count
                                    count={item.summary.critical}
                                    tone="text-red-700 dark:text-red-400"
                                />
                            </TableCell>
                            <TableCell className="text-right">
                                <Count
                                    count={item.summary.high}
                                    tone="text-orange-700 dark:text-orange-400"
                                />
                            </TableCell>
                            <TableCell className="text-right">
                                <Count
                                    count={item.summary.medium}
                                    tone="text-amber-700 dark:text-amber-400"
                                />
                            </TableCell>
                            <TableCell className="text-right tabular-nums">{item.summary.ignored}</TableCell>
                            <TableCell className="text-xs text-muted-foreground">
                                {timeAgoFormatter.format(item.scannedAt)}
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}
