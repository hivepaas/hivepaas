import { useState } from "react";

import { cn } from "@/lib/utils";
import { AlertTriangle } from "lucide-react";
import { SystemImageScanQueries } from "~/system-settings/data";
import type { ImageScanApp } from "~/system-settings/domain";

import { AppLink, AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { TablePagination } from "@/components/ui/table-pagination";

const PAGE_SIZE = 25;

function Count({ count, tone }: { count: { total: number; fixable: number }; tone: string }) {
    return (
        <span className="tabular-nums">
            <span className={cn(count.fixable > 0 && tone, count.fixable > 0 && "font-semibold")}>{count.fixable}</span>
            <span className="text-muted-foreground"> / {count.total}</span>
        </span>
    );
}

function ImagesCell({ item }: { item: ImageScanApp }) {
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
 * Every scanned app and what its images carry, the worst first: fixable
 * CRITICAL, then fixable HIGH. Counts are after ignores, fixable / all.
 */
export function SystemSettingsImageScanningAppsRoute() {
    const [pageIndex, setPageIndex] = useState(0);
    const { data, isLoading } = SystemImageScanQueries.useFindApps({
        pageOffset: pageIndex * PAGE_SIZE,
        pageLimit: PAGE_SIZE,
    });

    if (isLoading || !data) {
        return <AppLoader />;
    }
    if (data.data.length === 0) {
        return (
            <div className="rounded-md border px-4 py-8 text-center text-sm text-muted-foreground">
                No app has been scanned yet.
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-3">
            <p className="text-[13px] text-muted-foreground">
                Counts are fixable / all, after ignores. An app is scanned after each deployment, every day, and when
                asked on its Security tab.
            </p>
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
                        {data.data.map(item => (
                            <TableRow
                                key={item.app.id}
                                data-testid="image-scan-app-row"
                            >
                                <TableCell>
                                    <div className="flex flex-col gap-0.5">
                                        <AppLink.Basic
                                            to={ROUTE.projects.single.apps.single.security.$route(
                                                item.project.id,
                                                item.projectEnv.name,
                                                item.app.id,
                                            )}
                                            className="font-medium underline-offset-4 hover:underline"
                                        >
                                            {item.app.name}
                                        </AppLink.Basic>
                                        <span className="text-xs text-muted-foreground">
                                            {item.project.name} / {item.projectEnv.name}
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
            {data.total > PAGE_SIZE && (
                <TablePagination
                    pageIndex={pageIndex}
                    pageSize={PAGE_SIZE}
                    pageCount={Math.ceil(data.total / PAGE_SIZE)}
                    totalCount={data.total}
                    onPageChange={setPageIndex}
                    showPageSizeSelector={false}
                />
            )}
        </div>
    );
}
