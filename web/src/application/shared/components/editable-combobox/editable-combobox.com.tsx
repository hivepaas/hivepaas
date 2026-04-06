"use client";

import * as React from "react";

import { cn } from "@/lib/utils";
import { Check, X } from "lucide-react";

import { Command, CommandEmpty, CommandGroup, CommandItem, CommandList } from "@/components/ui/command";
import { Input } from "@/components/ui/input";
import { Popover, PopoverAnchor, PopoverContent } from "@/components/ui/popover";

export interface EditableComboboxProps {
    options: string[];
    value: string;
    onChange: (value: string) => void;
    placeholder?: string;
    className?: string;
    emptyText?: string;
    allowClear?: boolean;
}

export function EditableCombobox({
    options,
    value,
    onChange,
    placeholder = "Type or select...",
    className,
    emptyText = "No matching options",
    allowClear = true,
}: EditableComboboxProps) {
    const [open, setOpen] = React.useState(false);
    const inputRef = React.useRef<HTMLInputElement>(null);

    const filtered = React.useMemo(() => {
        if (!value) return options;
        const lower = value.toLowerCase();
        return options.filter(opt => opt.toLowerCase().includes(lower));
    }, [options, value]);

    const handleSelect = (selected: string) => {
        onChange(selected);
        setOpen(false);
        inputRef.current?.focus();
    };

    const showClear = allowClear && value.length > 0;

    return (
        <Popover
            open={open}
            onOpenChange={setOpen}
        >
            <PopoverAnchor asChild>
                <div className={cn("group/clear relative flex items-center", className)}>
                    <Input
                        ref={inputRef}
                        value={value}
                        onChange={e => {
                            onChange(e.target.value);
                        }}
                        onFocus={() => {
                            setOpen(true);
                        }}
                        placeholder={placeholder}
                        className={cn("h-9 w-full border-0", showClear && "pr-7")}
                    />
                    {showClear && (
                        <button
                            type="button"
                            tabIndex={-1}
                            aria-label="Clear"
                            className="absolute right-1.5 hidden rounded-sm p-0.5 text-muted-foreground hover:text-foreground group-hover/clear:inline-flex"
                            onPointerDown={e => {
                                e.preventDefault();
                            }}
                            onClick={() => {
                                onChange("");
                                inputRef.current?.focus();
                            }}
                        >
                            <X className="size-3.5" />
                        </button>
                    )}
                </div>
            </PopoverAnchor>
            <PopoverContent
                className="w-(--radix-popover-anchor-width) p-0"
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
                                                value === option ? "opacity-100" : "opacity-0",
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
    );
}
