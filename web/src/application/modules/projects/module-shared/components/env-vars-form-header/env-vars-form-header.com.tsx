import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { Input } from "@components/ui/input";
import {
    AArrowUp,
    ArrowDownAZ,
    ArrowUpAZ,
    EyeIcon,
    EyeOffIcon,
    LayoutList,
    SearchIcon,
    TableCellsMerge,
} from "lucide-react";
import { useDebounce } from "react-use";

function View({
    search,
    onRevealToggle,
    isRevealed,
    viewMode,
    onViewModeChange,
    onSortCycle,
    sortOrder = "normal",
    readOnly,
}: Props) {
    const [internalSearch, setInternalSearch] = useState(search?.value ?? "");

    useDebounce(
        () => {
            search?.onChange(internalSearch.trim());
        },
        350,
        [internalSearch],
    );

    return (
        <div className="sticky top-0 z-20 flex flex-col sm:flex-row sm:items-center justify-between gap-2.5 sm:gap-4 bg-background py-2.5">
            {/* Search Input */}
            <div className="relative w-full sm:w-64 sm:max-w-xs">
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
                    className="peer px-9 h-8 sm:h-9 text-xs sm:text-sm w-full [&::-webkit-search-cancel-button]:appearance-none [&::-webkit-search-decoration]:appearance-none [&::-webkit-search-results-button]:appearance-none [&::-webkit-search-results-decoration]:appearance-none"
                />
            </div>

            {/* Action Buttons */}
            <div className="flex items-center justify-start sm:justify-end gap-1.5 sm:gap-2 overflow-x-auto pb-0.5 sm:pb-0 scrollbar-none w-full sm:w-auto">
                {/* Sort Button */}
                {onSortCycle && (
                    <Button
                        type="button"
                        variant={sortOrder !== "normal" ? "default" : "outline"}
                        size="sm"
                        onClick={onSortCycle}
                        disabled={readOnly}
                        className="gap-1.5 shrink-0 px-2.5 sm:px-3 text-xs sm:text-sm h-8 sm:h-9"
                    >
                        {sortOrder === "asc" ? (
                            <>
                                <ArrowDownAZ className="size-3.5 sm:size-4" />
                                <span>A → Z</span>
                            </>
                        ) : sortOrder === "desc" ? (
                            <>
                                <ArrowUpAZ className="size-3.5 sm:size-4" />
                                <span>Z → A</span>
                            </>
                        ) : (
                            <>
                                <AArrowUp className="size-3.5 sm:size-4" />
                                <span>Sort</span>
                            </>
                        )}
                    </Button>
                )}

                {/* Reveal/Hide Toggle */}
                <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={onRevealToggle}
                    className="gap-1.5 shrink-0 px-2.5 sm:px-3 text-xs sm:text-sm h-8 sm:h-9"
                >
                    {isRevealed ? (
                        <>
                            <EyeOffIcon className="size-3.5 sm:size-4" />
                            <span>Hide</span>
                        </>
                    ) : (
                        <>
                            <EyeIcon className="size-3.5 sm:size-4" />
                            <span>Reveal</span>
                        </>
                    )}
                </Button>

                {/* Merge View Button */}
                <Button
                    type="button"
                    variant={viewMode === "merge" ? "default" : "outline"}
                    size="sm"
                    onClick={() => onViewModeChange?.("merge")}
                    className="gap-1.5 shrink-0 px-2.5 sm:px-3 text-xs sm:text-sm h-8 sm:h-9"
                >
                    <TableCellsMerge className="size-3.5 sm:size-4" />
                    <span className="hidden sm:inline">Merge View</span>
                    <span className="inline sm:hidden">Merge</span>
                </Button>

                {/* Individual View Button */}
                <Button
                    type="button"
                    variant={viewMode === "individual" ? "default" : "outline"}
                    size="sm"
                    onClick={() => onViewModeChange?.("individual")}
                    className="gap-1.5 shrink-0 px-2.5 sm:px-3 text-xs sm:text-sm h-8 sm:h-9"
                >
                    <LayoutList className="size-3.5 sm:size-4" />
                    <span className="hidden sm:inline">Individual View</span>
                    <span className="inline sm:hidden">Individual</span>
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
    onSortCycle?: () => void;
    sortOrder?: "normal" | "asc" | "desc";
    readOnly?: boolean;
};

export const EnvVarsFormHeader = React.memo(View);
