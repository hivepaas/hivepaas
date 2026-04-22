"use client";

import * as React from "react";

import { cn } from "@/lib/utils";
import { Check, RefreshCw, X } from "lucide-react";

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
    "allowClear"?: boolean;
    "aria-invalid"?: boolean;
    "onRefresh"?: () => void;
    "isRefreshing"?: boolean;
    "inputClassName"?: string;
    "disableFilter"?: boolean;
}

export function EditableCombobox({
    options,
    value,
    onChange,
    onInputChange,
    placeholder = "Type or select...",
    className,
    emptyText = "No matching options",
    allowClear = true,
    "aria-invalid": ariaInvalid,
    onRefresh,
    inputClassName,
    isRefreshing = false,
    disableFilter = false,
}: EditableComboboxProps) {
    const [open, setOpen] = React.useState(false);
    const inputRef = React.useRef<HTMLInputElement>(null);
    const text = value ?? "";
    const showRefresh = Boolean(onRefresh);

    const filtered = React.useMemo(() => {
        if (disableFilter || !text) return options;
        const lower = text.toLowerCase();
        return options.filter(opt => opt.toLowerCase().includes(lower));
    }, [options, text, disableFilter]);

    const handleSelect = (selected: string) => {
        onChange(selected);
        setOpen(false);
        inputRef.current?.focus();
    };

    const showClear = allowClear && text.length > 0;

    return (
        <div className={cn("flex w-full min-w-0 items-center gap-1.5", className)}>
            <div className="group/clear min-w-0 flex-1">
                <Popover
                    open={open}
                    onOpenChange={setOpen}
                >
                    <PopoverAnchor asChild>
                        <div className="relative flex w-full items-center">
                            <Input
                                ref={inputRef}
                                value={text}
                                onChange={e => {
                                    if (onInputChange) {
                                        onInputChange(e.target.value);
                                        return;
                                    }
                                    onChange(e.target.value);
                                }}
                                onFocus={() => {
                                    setOpen(true);
                                }}
                                placeholder={placeholder}
                                aria-invalid={ariaInvalid}
                                className={cn(
                                    "h-auto min-h-9 w-full min-w-0 overflow-hidden py-2 pr-8 pl-3 text-left font-normal leading-[18px]",
                                    showClear && "pr-9",
                                    inputClassName,
                                )}
                            />
                            {showClear && (
                                <button
                                    type="button"
                                    tabIndex={-1}
                                    aria-label="Clear"
                                    className="absolute right-2 hidden rounded-sm p-0.5 text-muted-foreground hover:text-foreground group-hover/clear:inline-flex"
                                    onPointerDown={e => {
                                        e.preventDefault();
                                    }}
                                    onClick={() => {
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
                        </div>
                    </PopoverAnchor>
                    <PopoverContent
                        className="w-fit p-0"
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
                                <CommandEmpty className="p-2 text-sm text-gray-500">{emptyText}</CommandEmpty>
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
                        onRefresh?.();
                    }}
                >
                    <RefreshCw className={cn("size-4", isRefreshing && "animate-spin")} />
                </Button>
            )}
        </div>
    );
}
