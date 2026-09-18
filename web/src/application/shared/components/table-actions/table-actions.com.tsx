import React, { type PropsWithChildren, useEffect, useRef, useState } from "react";

import { Input } from "@components/ui/input";
import { SearchIcon, X } from "lucide-react";

export function TableActions({ children, search, renderActions = null, renderAfterSearch = null }: Props) {
    const [internalSearch, setInternalSearch] = useState(search?.value ?? "");
    const inputRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        if (search?.autoFocus) {
            inputRef.current?.focus();
        }
    }, [search?.autoFocus]);

    useEffect(() => {
        setInternalSearch(search?.value ?? "");
    }, [search?.value]);

    const handleSearch = () => {
        search?.onChange(internalSearch.trim());
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === "Enter") {
            e.preventDefault();
            handleSearch();
        }
    };

    const handleClear = () => {
        setInternalSearch("");
        search?.onChange("");
        inputRef.current?.focus();
    };

    const placeholderText = search?.placeholder
        ? search.placeholder.toLowerCase().includes("enter")
            ? search.placeholder
            : `${search.placeholder} (press Enter)`
        : "Search (press Enter)...";

    return (
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
            <div className="flex flex-1 items-center gap-2 min-w-0">
                {children}

                {search && (
                    <div className="relative w-full sm:w-64 max-w-full">
                        <button
                            type="button"
                            onClick={handleSearch}
                            className="text-muted-foreground hover:text-foreground absolute inset-y-0 left-0 flex items-center justify-center pl-3 transition-colors cursor-pointer"
                            title="Search (press Enter)"
                            aria-label="Search"
                        >
                            <SearchIcon className="size-4" />
                        </button>
                        <Input
                            ref={inputRef}
                            value={internalSearch}
                            onChange={e => {
                                setInternalSearch(e.target.value);
                            }}
                            onKeyDown={handleKeyDown}
                            type="search"
                            placeholder={placeholderText}
                            className="peer pl-9 pr-9 [&::-webkit-search-cancel-button]:appearance-none [&::-webkit-search-decoration]:appearance-none [&::-webkit-search-results-button]:appearance-none [&::-webkit-search-results-decoration]:appearance-none w-full text-xs sm:text-sm"
                        />
                        {internalSearch && (
                            <button
                                type="button"
                                onClick={handleClear}
                                className="text-muted-foreground hover:text-foreground absolute inset-y-0 right-0 flex items-center justify-center pr-3 transition-colors cursor-pointer"
                                title="Clear search"
                                aria-label="Clear search"
                            >
                                <X className="size-3.5" />
                            </button>
                        )}
                    </div>
                )}

                {renderAfterSearch}
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
    renderAfterSearch?: React.ReactNode;
}>;
