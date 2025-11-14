import React, { type PropsWithChildren, useState } from "react";

import { Input } from "@components/ui/input";
import { SearchIcon } from "lucide-react";
import { useDebounce } from "react-use";

export function TableActions({ children, search, renderActions = null }: Props) {
    const [internalSearch, setInternalSearch] = useState(search?.value ?? "");

    useDebounce(
        () => {
            search?.onChange(internalSearch.trim());
        },
        350,
        [internalSearch],
    );

    return (
        <div className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-2">
                {children}

                {search && (
                    <div className="relative">
                        <div className="text-muted-foreground pointer-events-none absolute inset-y-0 left-0 flex items-center justify-center pl-3 peer-disabled:opacity-50">
                            <SearchIcon className="size-4" />
                            <span className="sr-only">Search</span>
                        </div>
                        <Input
                            value={internalSearch}
                            onChange={e => {
                                setInternalSearch(e.target.value);
                            }}
                            type="search"
                            placeholder="Search..."
                            className="peer px-9 [&::-webkit-search-cancel-button]:appearance-none [&::-webkit-search-decoration]:appearance-none [&::-webkit-search-results-button]:appearance-none [&::-webkit-search-results-decoration]:appearance-none"
                        />
                    </div>
                )}
            </div>

            {renderActions && <div className="flex items-center gap-2">{renderActions}</div>}
        </div>
    );
}

type Props = PropsWithChildren<{
    search?: {
        value: string;
        onChange: (search: string) => void;
    };
    renderActions?: React.ReactNode;
}>;
