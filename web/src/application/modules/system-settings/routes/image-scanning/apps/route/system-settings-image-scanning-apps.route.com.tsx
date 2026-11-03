import { useState } from "react";

import { ScannedAppsTable } from "~/projects/module-shared/components";
import { SystemImageScanQueries } from "~/system-settings/data";

import { AppLoader } from "@application/shared/components";

import { TablePagination } from "@/components/ui/table-pagination";

const PAGE_SIZE = 25;

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
            <ScannedAppsTable
                items={data.data}
                withProject
            />
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
