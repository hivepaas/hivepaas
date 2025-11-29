import * as React from "react";

import { cn } from "@/lib/utils";

import {
    Pagination,
    PaginationContent,
    PaginationEllipsis,
    PaginationItem,
    PaginationLink,
    PaginationNext,
    PaginationPrevious,
} from "./pagination";

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
    maxPageButtons?: number;
}

// Helper function to generate page numbers with ellipsis
function generatePageNumbers(currentPage: number, totalPages: number, maxButtons: number = 7): (number | "ellipsis")[] {
    if (totalPages <= maxButtons) {
        return Array.from({ length: totalPages }, (_, i) => i);
    }

    const pages: (number | "ellipsis")[] = [];
    const sideButtons = Math.floor((maxButtons - 3) / 2); // 3 for first, last, and one ellipsis

    // Always show first page
    pages.push(0);

    if (currentPage <= sideButtons + 1) {
        // Near the start
        for (let i = 1; i < maxButtons - 2; i++) {
            pages.push(i);
        }
        pages.push("ellipsis");
        pages.push(totalPages - 1);
    } else if (currentPage >= totalPages - sideButtons - 2) {
        // Near the end
        pages.push("ellipsis");
        for (let i = totalPages - maxButtons + 2; i < totalPages; i++) {
            pages.push(i);
        }
    } else {
        // In the middle
        pages.push("ellipsis");
        for (let i = currentPage - sideButtons; i <= currentPage + sideButtons; i++) {
            pages.push(i);
        }
        pages.push("ellipsis");
        pages.push(totalPages - 1);
    }

    return pages;
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
    maxPageButtons = 7,
}: TablePaginationProps) {
    const canPreviousPage = pageIndex > 0;
    const canNextPage = pageIndex < pageCount - 1;

    const startRow = totalCount && totalCount > 0 ? pageIndex * pageSize + 1 : 0;
    const endRow = Math.min((pageIndex + 1) * pageSize, totalCount ?? (pageIndex + 1) * pageSize);

    const pageNumbers = React.useMemo(
        () => generatePageNumbers(pageIndex, pageCount, maxPageButtons),
        [pageIndex, pageCount, maxPageButtons],
    );

    return (
        <div className={cn("flex items-center justify-between px-2 py-4", className)}>
            <div className="flex items-center gap-4">
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

            <Pagination>
                <PaginationContent>
                    <PaginationItem>
                        <PaginationPrevious
                            href="#"
                            onClick={e => {
                                e.preventDefault();
                                if (canPreviousPage) {
                                    onPageChange(pageIndex - 1);
                                }
                            }}
                            className={cn(!canPreviousPage && "pointer-events-none opacity-50")}
                        />
                    </PaginationItem>

                    {pageNumbers.map((page, index) => (
                        <PaginationItem key={`page-${index}`}>
                            {page === "ellipsis" ? (
                                <PaginationEllipsis />
                            ) : (
                                <PaginationLink
                                    href="#"
                                    onClick={e => {
                                        e.preventDefault();
                                        onPageChange(page);
                                    }}
                                    isActive={pageIndex === page}
                                >
                                    {page + 1}
                                </PaginationLink>
                            )}
                        </PaginationItem>
                    ))}

                    <PaginationItem>
                        <PaginationNext
                            href="#"
                            onClick={e => {
                                e.preventDefault();
                                if (canNextPage) {
                                    onPageChange(pageIndex + 1);
                                }
                            }}
                            className={cn(!canNextPage && "pointer-events-none opacity-50")}
                        />
                    </PaginationItem>
                </PaginationContent>
            </Pagination>
        </div>
    );
}

export { TablePagination };
