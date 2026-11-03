import { useCallback, useState } from "react";

import { type OrderState, type PaginationState, type RowSelectionState, type SortingState } from "@infrastructure/data";

interface TableStateDefaults {
    rowSelection?: RowSelectionState;
    pagination?: PaginationState;
    sorting?: SortingState;
    order?: OrderState;
    search?: string;
}

function createHook() {
    return function useTableState(defaults: TableStateDefaults = {}) {
        /**
         * Row selection state.
         */
        const [rowSelection, setRowSelection] = useState<RowSelectionState>(() => defaults.rowSelection ?? {});

        /**
         * Pagination state
         */
        const [pagination, setPagination] = useState<PaginationState>(() => ({
            page: 1,
            size: 10,

            ...defaults.pagination,
        }));

        /**
         * Sorting state
         */
        const [sorting, setSorting] = useState<SortingState>(() => defaults.sorting ?? []);

        /**
         * Order state
         */
        const [order, setOrder] = useState<OrderState>(() => defaults.order ?? []);

        /**
         * Search
         */
        const [search, setSearch] = useState(() => defaults.search ?? "");

        const handleSearchChange = useCallback((value: string) => {
            setSearch(value);

            setPagination(prev => ({ ...prev, page: 1 }));
        }, []);

        return {
            rowSelection,
            pagination,
            sorting,
            order,
            search,
            setRowSelection,
            setPagination,
            setSorting,
            setOrder,
            setSearch: handleSearchChange,
        };
    };
}

export const useTableState = createHook();
