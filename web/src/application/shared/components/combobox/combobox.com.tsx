"use client";

import * as React from "react";

import { cn } from "@/lib/utils";
import { Check, ChevronsUpDown, Loader2, RefreshCw, X } from "lucide-react";

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
    /** If true, disable the clear button on hover and deselect on click */
    "disableClear"?: boolean;
    "renderSelectedOption"?: (option: ComboboxOption<T>) => React.ReactNode;
    "renderOption"?: (option: ComboboxOption<T>) => React.ReactNode;
}

function getOptionValueString<T extends Record<string, unknown>>(option: ComboboxOption<T>, valueKey: keyof T): string {
    const raw = option.value[valueKey];
    if (typeof raw === "string") {
        return raw;
    }
    if (typeof raw === "number" || typeof raw === "boolean") {
        return `${raw}`;
    }
    return "";
}

function normalizeValue(value: string | number | null | undefined): string | null {
    if (value == null || value === "") {
        return null;
    }
    return (typeof value === "string" ? value : `${value}`).toLowerCase();
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
    disableClear = false,
    renderSelectedOption,
    renderOption,
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

    const normalizedValue = normalizeValue(value);

    const selectedOption = React.useMemo(() => {
        if (normalizedValue == null) return undefined;
        return options.find(opt => getOptionValueString(opt, valueKey).toLowerCase() === normalizedValue);
    }, [options, normalizedValue, valueKey]);

    const handleOpenChange = (nextOpen: boolean) => {
        setSearch("");
        setOpen(nextOpen);
    };

    const handleSelectOption = (option: ComboboxOption<T>) => {
        if (option.disabled) {
            return;
        }

        const optionValue = getOptionValueString(option, valueKey);
        const isDeselecting = !disableClear && normalizedValue != null && normalizedValue === optionValue.toLowerCase();
        const newValue = isDeselecting ? null : optionValue;
        const selectedOptionData = isDeselecting ? null : option.value;

        onChange?.(newValue, selectedOptionData);

        if (closeOnSelect) {
            setSearch("");
            setOpen(false);
        }
    };

    const showClear = !disableClear && !disabled && !loading && value != null && value !== "";

    const handleClear = (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        onChange?.(null, null);
        setSearch("");
        setOpen(false);
    };

    return (
        <div className={cn("flex w-full min-w-0 items-center gap-1.5", className)}>
            <div className="relative group/clear min-w-0 flex-1">
                <Popover
                    open={open}
                    onOpenChange={handleOpenChange}
                >
                    <PopoverTrigger asChild>
                        <Button
                            type="button"
                            variant="outline"
                            role="combobox"
                            aria-expanded={open}
                            aria-invalid={ariaInvalid}
                            disabled={disabled || loading}
                            className="h-auto min-h-9 w-full min-w-0 justify-between gap-2 overflow-hidden py-2"
                        >
                            <span className="min-w-0 flex-1 overflow-hidden text-left font-normal leading-[18px]">
                                {selectedOption ? (
                                    renderSelectedOption ? (
                                        renderSelectedOption(selectedOption)
                                    ) : (
                                        <ComboboxLabelContent
                                            label={selectedOption.label}
                                            splitBadge={splitLabelBadge}
                                        />
                                    )
                                ) : (
                                    <span className="text-muted-foreground">{placeholder}</span>
                                )}
                            </span>
                            {loading ? (
                                <Loader2 className="ml-2 h-4 w-4 shrink-0 animate-spin opacity-50" />
                            ) : (
                                <ChevronsUpDown
                                    className={cn(
                                        "ml-2 h-4 w-4 shrink-0 opacity-50",
                                        showClear && "group-hover/clear:opacity-0 transition-opacity",
                                    )}
                                />
                            )}
                        </Button>
                    </PopoverTrigger>
                    <PopoverContent
                        className="w-fit p-0"
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
                                        <CommandEmpty className="p-2 text-sm text-gray-500">{emptyText}</CommandEmpty>
                                        {options.length > 0 && (
                                            <CommandGroup>
                                                {options.map(option => {
                                                    const optionValue = getOptionValueString(option, valueKey);
                                                    const isChecked =
                                                        normalizedValue != null &&
                                                        normalizedValue === optionValue.toLowerCase();
                                                    return (
                                                        <CommandItem
                                                            key={optionValue}
                                                            value={
                                                                option.label
                                                                    ? `${option.label} ${optionValue}`
                                                                    : optionValue
                                                            }
                                                            disabled={option.disabled}
                                                            onSelect={() => {
                                                                handleSelectOption(option);
                                                            }}
                                                        >
                                                            {renderOption ? (
                                                                renderOption(option)
                                                            ) : (
                                                                <ComboboxLabelContent
                                                                    label={option.label}
                                                                    splitBadge={splitLabelBadge}
                                                                />
                                                            )}
                                                            <Check
                                                                className={cn(
                                                                    "ml-auto h-4 w-4 shrink-0",
                                                                    isChecked ? "opacity-100" : "opacity-0",
                                                                )}
                                                            />
                                                        </CommandItem>
                                                    );
                                                })}
                                            </CommandGroup>
                                        )}
                                    </>
                                )}
                            </CommandList>
                        </Command>
                    </PopoverContent>
                </Popover>

                {showClear && (
                    <button
                        type="button"
                        tabIndex={-1}
                        aria-label="Clear"
                        className="absolute right-2 top-1/2 -translate-y-1/2 hidden group-hover/clear:flex items-center justify-center size-5 rounded-sm text-muted-foreground hover:text-foreground bg-background hover:bg-muted transition-colors z-10"
                        onClick={handleClear}
                    >
                        <X className="size-3.5" />
                    </button>
                )}
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
