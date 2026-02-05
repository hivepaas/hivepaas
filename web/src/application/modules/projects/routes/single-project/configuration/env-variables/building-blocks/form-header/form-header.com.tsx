import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { Input } from "@components/ui/input";
import { EyeIcon, EyeOffIcon, LayoutList, SearchIcon, TableCellsMerge } from "lucide-react";
import { useDebounce } from "react-use";

function View({ search, onRevealToggle, isRevealed, viewMode, onViewModeChange }: Props) {
    const [internalSearch, setInternalSearch] = useState(search?.value ?? "");

    useDebounce(
        () => {
            search?.onChange(internalSearch.trim());
        },
        350,
        [internalSearch],
    );

    return (
        <div className="flex items-center justify-between gap-4">
            {/* Search Input */}
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
                    placeholder="Search"
                    className="peer px-9 [&::-webkit-search-cancel-button]:appearance-none [&::-webkit-search-decoration]:appearance-none [&::-webkit-search-results-button]:appearance-none [&::-webkit-search-results-decoration]:appearance-none"
                />
            </div>

            {/* Action Buttons */}
            <div className="flex items-center gap-2">
                {/* Reveal/Hide Toggle */}
                <Button
                    type="button"
                    variant="outline"
                    onClick={onRevealToggle}
                    className="gap-2"
                >
                    {isRevealed ? (
                        <>
                            <EyeOffIcon className="size-4" />
                            <span>Hide</span>
                        </>
                    ) : (
                        <>
                            <EyeIcon className="size-4" />
                            <span>Reveal</span>
                        </>
                    )}
                </Button>

                {/* Merge View Button */}
                <Button
                    type="button"
                    variant={viewMode === "merge" ? "default" : "outline"}
                    onClick={() => onViewModeChange?.("merge")}
                    className="gap-2"
                >
                    <TableCellsMerge className="size-4" />
                    <span>Merge View</span>
                </Button>

                {/* Individual View Button */}
                <Button
                    type="button"
                    variant={viewMode === "individual" ? "default" : "outline"}
                    onClick={() => onViewModeChange?.("individual")}
                    className="gap-2"
                >
                    <LayoutList className="size-4" />
                    <span>Individual View</span>
                </Button>
            </div>
        </div>
    );
}

type Props = {
    search?: {
        value: string;
        onChange: (search: string) => void;
    };
    isRevealed: boolean;
    onRevealToggle: () => void;
    viewMode: "merge" | "individual";
    onViewModeChange?: (mode: "merge" | "individual") => void;
};

export const FormHeader = React.memo(View);
