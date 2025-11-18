"use client";

import * as React from "react";

import { cn } from "@/lib/utils";
import { Check, ChevronsUpDown, Loader2 } from "lucide-react";

import { useDebouncedSearch } from "@application/shared/hooks";

import { Button } from "@/components/ui/button";
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

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
}: ComboboxProps<T>) {
    const [open, setOpen] = React.useState(false);
    const [debouncedSearch, setSearch, searchValue] = useDebouncedSearch(debounceMs, "");

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
                    className={cn("w-full justify-between", className)}
                >
                    {selectedOption ? selectedOption.label : placeholder}
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
                            placeholder="Search..."
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
                                                {option.label}
                                                <Check
                                                    className={cn(
                                                        "ml-auto h-4 w-4",
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
    );
}
