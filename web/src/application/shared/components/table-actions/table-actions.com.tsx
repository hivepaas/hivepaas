import React, { type PropsWithChildren, useEffect, useRef, useState } from "react";

import { Input } from "@components/ui/input";
import { SearchIcon } from "lucide-react";
import { useDebounce } from "react-use";

export function TableActions({ children, search, renderActions = null }: Props) {
    const [internalSearch, setInternalSearch] = useState(search?.value ?? "");
    const inputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        if (search?.autoFocus) {
            inputRef.current?.focus();
        }
    }, [search?.autoFocus]);

    useDebounce(
        () => {
            search?.onChange(internalSearch.trim());
        },
        350,
        [internalSearch],
    );

    return (
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
            <div className="flex flex-1 items-center gap-2 min-w-0">
                {children}

                {search && (
                    <div className="relative w-full sm:w-64 max-w-full">
                        <div className="text-muted-foreground pointer-events-none absolute inset-y-0 left-0 flex items-center justify-center pl-3 peer-disabled:opacity-50">
                            <SearchIcon className="size-4" />
                            <span className="sr-only">Search</span>
                        </div>
                        <Input
                            ref={inputRef}
                            value={internalSearch}
                            onChange={e => {
                                setInternalSearch(e.target.value);
                            }}
                            type="search"
                            placeholder={search.placeholder ?? "Search"}
                            className="peer px-9 [&::-webkit-search-cancel-button]:appearance-none [&::-webkit-search-decoration]:appearance-none [&::-webkit-search-results-button]:appearance-none [&::-webkit-search-results-decoration]:appearance-none w-full"
                        />
                    </div>
                )}
            </div>

            {renderActions && (
                <div className="flex flex-wrap items-center justify-start sm:justify-end gap-2 shrink-0">
                    {renderActions}
                </div>
            )}
        </div>
    );
}

type Props = PropsWithChildren<{
    search?: {
        value: string;
        onChange: (search: string) => void;
        autoFocus?: boolean;
        placeholder?: string;
    };
    renderActions?: React.ReactNode;
}>;
