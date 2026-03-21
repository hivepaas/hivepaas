"use client";

import * as React from "react";

import { cn } from "@/lib/utils";
import { Check, ChevronsUpDown, Loader2, RefreshCw } from "lucide-react";

import { useDebouncedSearch } from "@application/shared/hooks";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

function ComboboxLabelContent({ label, splitBadge }: { label: string; splitBadge: boolean }) {
    if (!splitBadge) {
        return <span className="truncate">{label}</span>;
    }
    const spaceIdx = label.indexOf(" ");
    if (spaceIdx <= 0) {
        return <span className="truncate">{label}</span>;
    }
    const first = label.slice(0, spaceIdx);
    const rest = label.slice(spaceIdx + 1).trimStart();
    return (
        <span className="flex min-w-0 max-w-full items-center gap-2 text-left">
            <Badge
                // variant="secondary"
                color="primary"
                className={cn(
                    "max-w-none shrink-0 rounded-md px-1.5 text-xs font-medium leading-none",
                    "overflow-visible",
                )}
            >
                {first}
            </Badge>
            <span className="min-w-0 flex-1 truncate font-normal">{rest}</span>
        </span>
    );
}

export interface ComboboxOption<T extends Record<string, unknown> = Record<string, unknown>> {
    value: T;
    label: string;
    disabled?: boolean;
}

export interface ComboboxProps<T extends Record<string, unknown> = Record<string, unknown>> {
    "options": ComboboxOption<T>[];
    "value"?: string | null;
    "onChange"?: (value: string | null, option: T | null) => void;
    "onSearch"?: (search: string) => void;
    "placeholder"?: string;
    "disabled"?: boolean;
    "searchable"?: boolean;
    "debounceMs"?: number;
    "className"?: string;
    "emptyText"?: string;
    "closeOnSelect"?: boolean;
    "valueKey"?: keyof T;
    "aria-invalid"?: boolean;
    "loading"?: boolean;
    /** Show refresh control; typically wire to query `refetch` */
    "onRefresh"?: () => void;
    /** Spin refresh icon while refetch is in flight */
    "isRefreshing"?: boolean;
    /** If label contains a space, render first token as a colored badge and the rest as text */
    "splitLabelBadge"?: boolean;
}

export function Combobox<T extends Record<string, unknown> = Record<string, unknown>>({
    options,
    value,
    onChange,
    onSearch,
    placeholder = "Select...",
    disabled = false,
    searchable = true,
    debounceMs = 250,
    className,
    emptyText = "No options available",
    closeOnSelect = true,
    valueKey = "id",
    "aria-invalid": ariaInvalid,
    loading = false,
    onRefresh,
    isRefreshing = false,
    splitLabelBadge = false,
}: ComboboxProps<T>) {
    const [open, setOpen] = React.useState(false);
    const [debouncedSearch, setSearch, searchValue] = useDebouncedSearch(debounceMs, "");

    const showRefresh = Boolean(onRefresh);

    // Call onSearch callback when debounced search changes
    React.useEffect(() => {
        if (onSearch) {
            onSearch(debouncedSearch);
        }
    }, [debouncedSearch, onSearch]);

    const selectedOption = React.useMemo(() => {
        return options.find(opt => String(opt.value[valueKey]) === String(value));
    }, [options, value, valueKey]);

    const handleSelect = (currentValue: string) => {
        const option = options.find(opt => String(opt.value[valueKey]) === currentValue);
        if (!option || option.disabled) {
            return;
        }

        const optionValue = String(option.value[valueKey]);
        const isDeselecting = currentValue === String(value);
        const newValue = isDeselecting ? null : optionValue;
        const selectedOptionData = isDeselecting ? null : option.value;

        onChange?.(newValue, selectedOptionData);

        if (closeOnSelect) {
            setSearch("");
            setOpen(false);
        }
    };

    return (
        <div className={cn("flex w-full items-center gap-1.5", className)}>
            <div className="min-w-0 flex-1">
                <Popover
                    open={open}
                    onOpenChange={setOpen}
                >
                    <PopoverTrigger asChild>
                        <Button
                            variant="outline"
                            role="combobox"
                            aria-expanded={open}
                            aria-invalid={ariaInvalid}
                            disabled={disabled || loading}
                            className="h-auto min-h-9 w-full justify-between gap-2 py-2"
                        >
                            <span className="min-w-0 flex-1 text-left font-normal">
                                {selectedOption ? (
                                    <ComboboxLabelContent
                                        label={selectedOption.label}
                                        splitBadge={splitLabelBadge}
                                    />
                                ) : (
                                    placeholder
                                )}
                            </span>
                            {loading ? (
                                <Loader2 className="ml-2 h-4 w-4 shrink-0 animate-spin opacity-50" />
                            ) : (
                                <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                            )}
                        </Button>
                    </PopoverTrigger>
                    <PopoverContent
                        className="w-(--radix-popover-trigger-width) p-0"
                        align="start"
                    >
                        <Command shouldFilter={!onSearch}>
                            {searchable && (
                                <CommandInput
                                    placeholder="Search"
                                    value={searchValue}
                                    onValueChange={setSearch}
                                />
                            )}
                            <CommandList>
                                {loading ? (
                                    <div className="flex items-center justify-center py-6">
                                        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                                    </div>
                                ) : (
                                    <>
                                        <CommandEmpty>{emptyText}</CommandEmpty>
                                        <CommandGroup>
                                            {options.map(option => {
                                                const optionValue = String(option.value[valueKey]);
                                                return (
                                                    <CommandItem
                                                        key={optionValue}
                                                        value={optionValue}
                                                        disabled={option.disabled}
                                                        onSelect={handleSelect}
                                                    >
                                                        <ComboboxLabelContent
                                                            label={option.label}
                                                            splitBadge={splitLabelBadge}
                                                        />
                                                        <Check
                                                            className={cn(
                                                                "ml-auto h-4 w-4 shrink-0",
                                                                value === optionValue ? "opacity-100" : "opacity-0",
                                                            )}
                                                        />
                                                    </CommandItem>
                                                );
                                            })}
                                        </CommandGroup>
                                    </>
                                )}
                            </CommandList>
                        </Command>
                    </PopoverContent>
                </Popover>
            </div>
            {showRefresh && (
                <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    disabled={disabled}
                    aria-label="Refresh list"
                    title="Refresh list"
                    className="size-9 shrink-0 shadow-none"
                    onClick={() => {
                        if (onRefresh) {
                            onRefresh();
                        }
                    }}
                >
                    <RefreshCw className={cn("size-4", isRefreshing && "animate-spin")} />
                </Button>
            )}
        </div>
    );
}
