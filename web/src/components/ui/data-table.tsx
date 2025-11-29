import * as React from "react";

import { cn } from "@/lib/utils";
import type { PaginationState } from "@infrastructure/data";
import {
    type ColumnDef,
    type ColumnFiltersState,
    type Header,
    type SortingState,
    type Table as TanstackTable,
    type VisibilityState,
    flexRender,
    getCoreRowModel,
    getFilteredRowModel,
    getPaginationRowModel,
    getSortedRowModel,
    useReactTable,
} from "@tanstack/react-table";
import { ArrowDown, ArrowUp, ArrowUpDown, Loader2 } from "lucide-react";

import { Button } from "./button";
import { Skeleton } from "./skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./table";
import { TablePagination } from "./table-pagination";

export type ColumnAlign = "left" | "center" | "right";
export type ColumnVerticalAlign = "top" | "middle" | "bottom";

export interface ColumnAlignmentMeta {
    align?: ColumnAlign;
    verticalAlign?: ColumnVerticalAlign;
    titleAlign?: ColumnAlign;
    sticky?: "left" | "right";
}

declare module "@tanstack/react-table" {
    interface ColumnMeta<TData, TValue> extends ColumnAlignmentMeta {}
}

interface DataTableColumnHeaderProps<TData, TValue> extends React.HTMLAttributes<HTMLDivElement> {
    header: Header<TData, TValue>;
    title: string;
}

// Helper functions to convert alignment to Tailwind classes
const getTextAlignClass = (align?: ColumnAlign): string => {
    switch (align) {
        case "center":
            return "text-center";
        case "right":
            return "text-right";
        case "left":
        default:
            return "text-left";
    }
};

const getVerticalAlignClass = (verticalAlign?: ColumnVerticalAlign): string => {
    switch (verticalAlign) {
        case "top":
            return "align-top";
        case "bottom":
            return "align-bottom";
        case "middle":
        default:
            return "align-middle";
    }
};

const getStickyClass = (sticky?: "left" | "right"): string => {
    if (!sticky) return "";
    if (sticky === "left") {
        return "sticky left-0 z-10";
    }
    return "sticky right-0 z-10";
};

function DataTableColumnHeader<TData, TValue>({ header, title, className }: DataTableColumnHeaderProps<TData, TValue>) {
    const column = header.column;
    const canSort = column.getCanSort();
    const isSorted = column.getIsSorted();
    const titleAlign = column.columnDef.meta?.titleAlign || column.columnDef.meta?.align;

    if (!canSort) {
        return <div className={cn(getTextAlignClass(titleAlign), className)}>{title}</div>;
    }

    const flexJustifyClass =
        titleAlign === "center" ? "justify-center" : titleAlign === "right" ? "justify-end" : "justify-start";

    return (
        <div className={cn("flex items-center space-x-2", flexJustifyClass, className)}>
            <Button
                variant="ghost"
                size="sm"
                className="-ml-3 h-8 data-[state=open]:bg-accent"
                onClick={() => column.toggleSorting(isSorted === "asc")}
            >
                <span>{title}</span>
                {isSorted === "desc" ? (
                    <ArrowDown className="ml-2 size-4" />
                ) : isSorted === "asc" ? (
                    <ArrowUp className="ml-2 size-4" />
                ) : (
                    <ArrowUpDown className="ml-2 size-4" />
                )}
            </Button>
        </div>
    );
}

// Helper functions to convert between codebase PaginationState and tanstack table format
const toTanstackPagination = (pagination: PaginationState) => ({
    pageIndex: pagination.page - 1, // Convert from 1-based to 0-based
    pageSize: pagination.size,
});

const fromTanstackPagination = (pageIndex: number, pageSize: number): PaginationState => ({
    page: pageIndex + 1, // Convert from 0-based to 1-based
    size: pageSize,
});

export interface DataTableProps<TData, TValue> {
    columns: ColumnDef<TData, TValue>[];
    data: TData[];
    pageSize?: number;
    pageSizeOptions?: number[];
    enablePagination?: boolean;
    enableSorting?: boolean;
    enableFiltering?: boolean;
    onRowClick?: (row: TData) => void;
    onSortingChange?: (sorting: SortingState) => void;
    onPaginationChange?: (pagination: PaginationState) => void;
    onTableChange?: (table: TanstackTable<TData>) => void;
    initialSorting?: SortingState;
    initialPagination?: PaginationState;
    className?: string;
    headerClassName?: string;
    bodyClassName?: string;
    rowClassName?: string | ((row: TData) => string);
    showTotalCount?: boolean;
    showPageSizeSelector?: boolean;
    totalCount?: number;
    manualPagination?: boolean;
    manualSorting?: boolean;
    manualFiltering?: boolean;
    isLoading?: boolean;
}

function DataTable<TData, TValue>({
    columns,
    data,
    pageSize = 10,
    pageSizeOptions = [10, 20, 30, 50, 100],
    enablePagination = true,
    enableSorting = true,
    enableFiltering = true,
    onRowClick,
    onSortingChange,
    onPaginationChange,
    onTableChange,
    initialSorting = [],
    initialPagination,
    className,
    headerClassName,
    bodyClassName,
    rowClassName,
    showTotalCount = true,
    showPageSizeSelector = true,
    totalCount,
    manualPagination = false,
    manualSorting = false,
    manualFiltering = false,
    isLoading = false,
}: DataTableProps<TData, TValue>) {
    const [sorting, setSorting] = React.useState<SortingState>(initialSorting);
    const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
    const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({});
    const [rowSelection, setRowSelection] = React.useState({});

    // Convert initialPagination to tanstack format
    const initialTanstackPagination = React.useMemo(() => {
        if (initialPagination) {
            return toTanstackPagination(initialPagination);
        }
        return { pageIndex: 0, pageSize };
    }, [initialPagination, pageSize]);

    const handleSortingChange = React.useCallback(
        (updater: SortingState | ((prev: SortingState) => SortingState)) => {
            const newSorting = typeof updater === "function" ? updater(sorting) : updater;
            setSorting(newSorting);
            onSortingChange?.(newSorting);
        },
        [sorting, onSortingChange],
    );

    const table = useReactTable({
        data,
        columns,
        getCoreRowModel: getCoreRowModel(),
        getPaginationRowModel: enablePagination && !manualPagination ? getPaginationRowModel() : undefined,
        getSortedRowModel: enableSorting && !manualSorting ? getSortedRowModel() : undefined,
        getFilteredRowModel: enableFiltering && !manualFiltering ? getFilteredRowModel() : undefined,
        onSortingChange: handleSortingChange,
        onColumnFiltersChange: setColumnFilters,
        onColumnVisibilityChange: setColumnVisibility,
        onRowSelectionChange: setRowSelection,
        manualPagination,
        manualSorting,
        manualFiltering,
        pageCount: manualPagination && totalCount ? Math.ceil(totalCount / pageSize) : undefined,
        state: {
            sorting,
            columnFilters,
            columnVisibility,
            rowSelection,
        },
        initialState: {
            pagination: initialTanstackPagination,
            sorting: initialSorting,
        },
    });

    React.useEffect(() => {
        onTableChange?.(table);
    }, [table, onTableChange]);

    const handlePageChange = React.useCallback(
        (pageIndex: number) => {
            table.setPageIndex(pageIndex);
            const currentPageSize = table.getState().pagination.pageSize;
            const paginationState = fromTanstackPagination(pageIndex, currentPageSize);
            onPaginationChange?.(paginationState);
        },
        [table, onPaginationChange],
    );

    const handlePageSizeChange = React.useCallback(
        (newPageSize: number) => {
            table.setPageSize(newPageSize);
            const newPageIndex = table.getState().pagination.pageIndex;
            const paginationState = fromTanstackPagination(newPageIndex, newPageSize);
            onPaginationChange?.(paginationState);
        },
        [table, onPaginationChange],
    );

    const getRowClassName = React.useCallback(
        (row: TData) => {
            if (typeof rowClassName === "function") {
                return rowClassName(row);
            }
            return rowClassName;
        },
        [rowClassName],
    );

    return (
        <div className={cn("space-y-4", className)}>
            <div className="relative rounded-md border overflow-hidden">
                {/* Progress bar */}
                {isLoading && (
                    <div className="absolute top-0 left-0 right-0 z-10 h-1 overflow-hidden rounded-t-md bg-muted">
                        <div
                            className="h-full w-1/3 bg-primary"
                            style={{
                                animation: "progress 1.5s ease-in-out infinite",
                            }}
                        />
                    </div>
                )}

                {/* Loading overlay */}
                {isLoading && (
                    <div className="absolute inset-0 z-20 flex items-center justify-center rounded-md bg-transparent">
                        <div className="flex flex-col items-center gap-2">
                            <div className="h-10 w-10 animate-spin rounded-full border-b-2 border-t-2 border-primary" />
                            <span className="text-sm text-muted-foreground">Loading...</span>
                        </div>
                    </div>
                )}

                <Table className={cn(isLoading && "opacity-50")}>
                    <TableHeader className={headerClassName}>
                        {table.getHeaderGroups().map(headerGroup => (
                            <TableRow key={headerGroup.id}>
                                {headerGroup.headers.map(header => {
                                    const canSort = header.column.getCanSort();
                                    const title =
                                        typeof header.column.columnDef.header === "string"
                                            ? header.column.columnDef.header
                                            : header.column.id;
                                    const meta = header.column.columnDef.meta;
                                    const titleAlign = meta?.titleAlign || meta?.align;
                                    const verticalAlign = meta?.verticalAlign;
                                    const sticky = meta?.sticky;

                                    return (
                                        <TableHead
                                            key={header.id}
                                            className={cn(
                                                canSort && enableSorting ? "cursor-pointer select-none" : "",
                                                getTextAlignClass(titleAlign),
                                                getVerticalAlignClass(verticalAlign),
                                                getStickyClass(sticky),
                                            )}
                                        >
                                            {header.isPlaceholder ? null : canSort && enableSorting ? (
                                                <DataTableColumnHeader
                                                    header={header}
                                                    title={title}
                                                />
                                            ) : (
                                                flexRender(header.column.columnDef.header, header.getContext())
                                            )}
                                        </TableHead>
                                    );
                                })}
                            </TableRow>
                        ))}
                    </TableHeader>
                    <TableBody className={bodyClassName}>
                        {isLoading ? (
                            // Show skeleton rows when loading
                            Array.from({ length: table.getRowModel().rows?.length ?? pageSize }).map((_, index) => (
                                <TableRow key={`skeleton-${index}`}>
                                    {columns.map((_, colIndex) => (
                                        <TableCell key={`skeleton-cell-${index}-${colIndex}`}>
                                            <Skeleton className="h-4 w-full" />
                                        </TableCell>
                                    ))}
                                </TableRow>
                            ))
                        ) : table.getRowModel().rows?.length ? (
                            table.getRowModel().rows.map(row => (
                                <TableRow
                                    key={row.id}
                                    data-state={row.getIsSelected() && "selected"}
                                    className={cn(onRowClick && "cursor-pointer", getRowClassName(row.original))}
                                    onClick={() => onRowClick?.(row.original)}
                                >
                                    {row.getVisibleCells().map(cell => {
                                        const meta = cell.column.columnDef.meta;
                                        const align = meta?.align;
                                        const verticalAlign = meta?.verticalAlign;
                                        const sticky = meta?.sticky;

                                        return (
                                            <TableCell
                                                key={cell.id}
                                                className={cn(
                                                    getTextAlignClass(align),
                                                    getVerticalAlignClass(verticalAlign),
                                                    getStickyClass(sticky),
                                                )}
                                            >
                                                {flexRender(cell.column.columnDef.cell, cell.getContext())}
                                            </TableCell>
                                        );
                                    })}
                                </TableRow>
                            ))
                        ) : (
                            <TableRow>
                                <TableCell
                                    colSpan={columns.length}
                                    className="h-24 text-center"
                                >
                                    No results.
                                </TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </div>
            {enablePagination && (
                <TablePagination
                    pageIndex={table.getState().pagination.pageIndex}
                    pageSize={table.getState().pagination.pageSize}
                    pageCount={table.getPageCount()}
                    totalCount={totalCount}
                    onPageChange={handlePageChange}
                    onPageSizeChange={handlePageSizeChange}
                    pageSizeOptions={pageSizeOptions}
                    showPageSizeSelector={showPageSizeSelector}
                    showTotalCount={showTotalCount}
                />
            )}
        </div>
    );
}

export { DataTable };
export type { TanstackTable, SortingState };
