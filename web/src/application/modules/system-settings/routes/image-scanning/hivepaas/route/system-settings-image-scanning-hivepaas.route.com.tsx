import { dashedBorderBox, inlineLink } from "@lib/styles";
import { cn } from "@lib/utils";
import { AlertTriangle, ArrowUp } from "lucide-react";
import { Link } from "react-router";
import type { ScannedComponent } from "~/system-settings/api/services";
import { HivePaaSUpdatesQueries, SystemImageScanQueries } from "~/system-settings/data";

import { AppLink, AppLoader } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

function Count({ count, tone }: { count: { total: number; fixable: number }; tone: string }) {
    return (
        <span className="tabular-nums">
            <span className={cn(count.fixable > 0 && tone, count.fixable > 0 && "font-semibold")}>{count.fixable}</span>
            <span className="text-muted-foreground"> / {count.total}</span>
        </span>
    );
}

/** The release HivePaaS can move to, when one is newer: it pins newer images. */
function NewerRelease() {
    const { data } = HivePaaSUpdatesQueries.useFindReleaseInfo({ retry: false }, { quiet: true });
    if (!data) {
        return null;
    }
    const { stable, beta } = data.data;
    const target = stable?.canUpdate ? stable : beta?.canUpdate ? beta : null;
    if (!target) {
        return null;
    }
    return (
        <div className="flex items-start gap-2 rounded-md border border-amber-300 bg-amber-50 px-3 py-2 text-[13px] text-amber-950 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-100">
            <ArrowUp className="mt-0.5 size-3.5 shrink-0" />
            <span>
                HivePaaS <span className="font-mono">{target.appVersion}</span> is out: its images are newer, and may
                carry the fixes.{" "}
                <Link
                    to={ROUTE.systemSettings.hivepaas.updates.$route}
                    className={inlineLink}
                >
                    Update HivePaaS
                </Link>
            </span>
        </div>
    );
}

function ImagesCell({ item }: { item: ScannedComponent }) {
    return (
        <div className="flex flex-col gap-0.5">
            {item.images.map(image => (
                <span
                    key={image.imageId}
                    className="font-mono text-xs break-all"
                    title={image.imageId}
                >
                    {image.imageRef || image.imageId.slice(0, 19)}
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
 * What HivePaaS's own parts run and carry, as the daily scan found them
 * running. They are fixed by updating HivePaaS: nothing here is notified, nor
 * in Needs attention.
 */
export function SystemSettingsImageScanningHivePaaSRoute() {
    const { data, isLoading } = SystemImageScanQueries.useFindComponents();

    if (isLoading || !data) {
        return <AppLoader />;
    }
    const { components } = data.data;

    return (
        <div className="flex flex-col gap-3">
            <div className={cn(dashedBorderBox, "text-[13px]")}>
                HivePaaS&apos;s own images, as the daily scan found them running. They are fixed by updating HivePaaS,
                not by anything rebuilt here: they are neither notified nor listed in Needs attention. Counts are
                fixable / all, after the global ignores.
            </div>
            <NewerRelease />
            {components.length === 0 ? (
                <div className="rounded-md border px-4 py-8 text-center text-sm text-muted-foreground">
                    Not scanned yet: the daily scan scans HivePaaS&apos;s images, and Actions runs it now.
                </div>
            ) : (
                <div className="overflow-x-auto rounded-md border">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Part</TableHead>
                                <TableHead>Images</TableHead>
                                <TableHead className="text-right">Critical</TableHead>
                                <TableHead className="text-right">High</TableHead>
                                <TableHead className="text-right">Medium</TableHead>
                                <TableHead className="text-right">Ignored</TableHead>
                                <TableHead>Scanned</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {components.map(item => (
                                <TableRow
                                    key={item.component.key}
                                    data-testid="image-scan-component-row"
                                >
                                    <TableCell>
                                        <AppLink.Basic
                                            to={ROUTE.systemSettings.imageScanning.hivepaas.single.$route(
                                                item.component.key,
                                            )}
                                            className="font-medium underline-offset-4 hover:underline"
                                        >
                                            {item.component.name}
                                        </AppLink.Basic>
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
            )}
        </div>
    );
}
