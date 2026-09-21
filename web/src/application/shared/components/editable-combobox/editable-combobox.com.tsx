"use client";

import * as React from "react";

import { cn } from "@/lib/utils";
import { Check, ChevronDown, RefreshCw, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Command, CommandEmpty, CommandGroup, CommandItem, CommandList } from "@/components/ui/command";
import { Input } from "@/components/ui/input";
import { Popover, PopoverAnchor, PopoverContent } from "@/components/ui/popover";

export interface EditableComboboxProps {
    "options": string[];
    "value"?: string | null;
    "onChange": (value: string) => void;
    "onInputChange"?: (value: string) => void;
    "placeholder"?: string;
    "className"?: string;
    "emptyText"?: string;
    "disableClear"?: boolean;
    "aria-invalid"?: boolean;
    "onRefresh"?: () => void;
    "isRefreshing"?: boolean;
    "inputClassName"?: string;
    "disableFilter"?: boolean;
    "disabled"?: boolean;
    "popoverClassName"?: string;
}

export function EditableCombobox({
    options = [],
    value,
    onChange,
    onInputChange,
    placeholder = "Type or select...",
    className,
    emptyText = "No matching options",
    disableClear = false,
    "aria-invalid": ariaInvalid,
    onRefresh,
    inputClassName,
    isRefreshing = false,
    disableFilter = false,
    disabled = false,
    popoverClassName,
}: EditableComboboxProps) {
    const [open, setOpen] = React.useState(false);
    const [search, setSearch] = React.useState("");
    const inputRef = React.useRef<HTMLInputElement>(null);
    const text = value ?? "";
    const showRefresh = Boolean(onRefresh);

    const safeOptions = React.useMemo(() => (Array.isArray(options) ? options : []), [options]);

    const filtered = React.useMemo(() => {
        if (disableFilter || !search) return safeOptions;
        const lower = search.toLowerCase();
        return safeOptions.filter(opt => opt && typeof opt === "string" && opt.toLowerCase().includes(lower));
    }, [safeOptions, search, disableFilter]);

    const updateOpen = (nextOpen: boolean) => {
        if (disabled) {
            setOpen(false);
            setSearch("");
            return;
        }

        if (nextOpen) {
            setSearch("");
        }

        setOpen(nextOpen);
    };

    const handleSelect = (selected: string) => {
        if (disabled) {
            return;
        }

        onChange(selected);
        setSearch("");
        setOpen(false);
        inputRef.current?.focus();
    };

    const showClear = !disableClear && !disabled && text.length > 0;

    return (
        <div className={cn("flex w-full min-w-0 items-center gap-1.5", className)}>
            <div className="group/clear min-w-0 flex-1">
                <Popover
                    open={!disabled && open}
                    onOpenChange={updateOpen}
                >
                    <PopoverAnchor asChild>
                        <div className="relative flex w-full items-center">
                            <Input
                                ref={inputRef}
                                value={text}
                                onChange={e => {
                                    if (disabled) {
                                        return;
                                    }

                                    setSearch(e.target.value);

                                    if (onInputChange) {
                                        onInputChange(e.target.value);
                                        return;
                                    }
                                    onChange(e.target.value);
                                }}
                                onFocus={() => {
                                    if (disabled) {
                                        return;
                                    }

                                    updateOpen(true);
                                }}
                                placeholder={placeholder}
                                aria-invalid={ariaInvalid}
                                className={cn(
                                    "h-auto min-h-9 w-full min-w-0 overflow-hidden py-2 pr-8 pl-3 text-left font-normal leading-[18px]",
                                    showClear && "pr-9",
                                    inputClassName,
                                )}
                                disabled={disabled}
                            />
                            <div className="absolute right-2 flex items-center gap-0.5">
                                {showClear && (
                                    <button
                                        type="button"
                                        tabIndex={-1}
                                        aria-label="Clear"
                                        className="hidden rounded-sm p-0.5 text-muted-foreground hover:text-foreground group-hover/clear:inline-flex"
                                        onPointerDown={e => {
                                            e.preventDefault();
                                        }}
                                        onClick={() => {
                                            setSearch("");
                                            if (onInputChange) {
                                                onInputChange("");
                                            } else {
                                                onChange("");
                                            }
                                            inputRef.current?.focus();
                                        }}
                                    >
                                        <X className="size-3.5" />
                                    </button>
                                )}
                                <button
                                    type="button"
                                    tabIndex={-1}
                                    aria-label="Toggle options"
                                    className={cn(
                                        "rounded-sm p-0.5 text-muted-foreground hover:text-foreground",
                                        showClear && "group-hover/clear:hidden",
                                    )}
                                    onClick={() => {
                                        if (disabled) return;
                                        updateOpen(!open);
                                        inputRef.current?.focus();
                                    }}
                                >
                                    <ChevronDown
                                        className={cn(
                                            "size-3.5 opacity-60 transition-transform duration-200",
                                            open && "rotate-180",
                                        )}
                                    />
                                </button>
                            </div>
                        </div>
                    </PopoverAnchor>
                    <PopoverContent
                        className={cn("w-fit min-w-[var(--radix-popover-anchor-width)] p-0", popoverClassName)}
                        align="start"
                        onOpenAutoFocus={e => {
                            e.preventDefault();
                        }}
                        onInteractOutside={e => {
                            if (e.target === inputRef.current) {
                                e.preventDefault();
                            }
                        }}
                    >
                        <Command shouldFilter={false}>
                            <CommandList>
                                <CommandEmpty className="p-2 text-sm text-muted-foreground">{emptyText}</CommandEmpty>
                                {filtered.length > 0 && (
                                    <CommandGroup>
                                        {filtered.map(option => (
                                            <CommandItem
                                                key={option}
                                                value={option}
                                                onSelect={handleSelect}
                                            >
                                                <span className="truncate">{option}</span>
                                                <Check
                                                    className={cn(
                                                        "ml-auto h-4 w-4 shrink-0",
                                                        text === option ? "opacity-100" : "opacity-0",
                                                    )}
                                                />
                                            </CommandItem>
                                        ))}
                                    </CommandGroup>
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
                    aria-label="Refresh list"
                    title="Refresh list"
                    className="size-9 shrink-0 shadow-none"
                    onClick={() => {
                        if (disabled) {
                            return;
                        }

                        onRefresh?.();
                    }}
                    disabled={disabled || isRefreshing}
                >
                    <RefreshCw className={cn("size-4", isRefreshing && "animate-spin")} />
                </Button>
            )}
        </div>
    );
}
