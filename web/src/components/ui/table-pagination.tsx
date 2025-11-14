import * as React from "react";

import { cn } from "@/lib/utils";
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from "lucide-react";

import { Button } from "./button";

export interface TablePaginationProps {
    pageIndex: number;
    pageSize: number;
    pageCount: number;
    totalCount?: number;
    onPageChange: (pageIndex: number) => void;
    onPageSizeChange?: (pageSize: number) => void;
    pageSizeOptions?: number[];
    className?: string;
    showPageSizeSelector?: boolean;
    showTotalCount?: boolean;
}

function TablePagination({
    pageIndex,
    pageSize,
    pageCount,
    totalCount,
    onPageChange,
    onPageSizeChange,
    pageSizeOptions = [10, 20, 30, 50, 100],
    className,
    showPageSizeSelector = true,
    showTotalCount = true,
}: TablePaginationProps) {
    const canPreviousPage = pageIndex > 0;
    const canNextPage = pageIndex < pageCount - 1;

    const startRow = pageIndex * pageSize + 1;
    const endRow = Math.min((pageIndex + 1) * pageSize, totalCount ?? (pageIndex + 1) * pageSize);

    return (
        <div className={cn("flex items-center justify-between px-2 py-4", className)}>
            <div className="flex items-center gap-2">
                {showPageSizeSelector && onPageSizeChange && (
                    <div className="flex items-center gap-2">
                        <span className="text-sm text-muted-foreground">Rows per page:</span>
                        <select
                            value={pageSize}
                            onChange={e => {
                                onPageSizeChange(Number(e.target.value));
                            }}
                            className="h-8 rounded-md border border-input bg-background px-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                        >
                            {pageSizeOptions.map(size => (
                                <option
                                    key={size}
                                    value={size}
                                >
                                    {size}
                                </option>
                            ))}
                        </select>
                    </div>
                )}
                {showTotalCount && totalCount !== undefined && (
                    <span className="text-sm text-muted-foreground">
                        {startRow}-{endRow} of {totalCount}
                    </span>
                )}
            </div>

            <div className="flex items-center gap-2">
                <div className="flex items-center gap-1">
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => onPageChange(0)}
                        disabled={!canPreviousPage}
                        aria-label="Go to first page"
                    >
                        <ChevronsLeft className="size-4" />
                    </Button>
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => onPageChange(pageIndex - 1)}
                        disabled={!canPreviousPage}
                        aria-label="Go to previous page"
                    >
                        <ChevronLeft className="size-4" />
                    </Button>
                    <div className="flex items-center gap-1 px-2">
                        <span className="text-sm font-medium">
                            Page {pageIndex + 1} of {pageCount || 1}
                        </span>
                    </div>
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => onPageChange(pageIndex + 1)}
                        disabled={!canNextPage}
                        aria-label="Go to next page"
                    >
                        <ChevronRight className="size-4" />
                    </Button>
                    <Button
                        variant="outline"
                        size="sm"
                        onClick={() => onPageChange(pageCount - 1)}
                        disabled={!canNextPage}
                        aria-label="Go to last page"
                    >
                        <ChevronsRight className="size-4" />
                    </Button>
                </div>
            </div>
        </div>
    );
}

export { TablePagination };
