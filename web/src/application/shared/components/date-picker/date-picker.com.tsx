import * as React from "react";

import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { CalendarIcon, X } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

export interface DatePickerProps {
    "value"?: Date | null;
    "onChange"?: (date: Date | undefined) => void;
    "placeholder"?: string;
    "disabled"?: boolean;
    "className"?: string;
    "aria-invalid"?: boolean;
    "allowClear"?: boolean;
}

export function DatePicker({
    value,
    onChange,
    placeholder = "Pick a date",
    disabled,
    className,
    "aria-invalid": ariaInvalid,
    allowClear = false,
}: DatePickerProps) {
    const showClear = allowClear && !disabled && Boolean(value);

    return (
        <Popover>
            <div className="group/clear">
                <PopoverTrigger asChild>
                    <Button
                        variant="outline"
                        disabled={disabled}
                        aria-invalid={ariaInvalid}
                        className={cn(
                            "w-full justify-between text-left font-normal",
                            !value && "text-muted-foreground",
                            className,
                        )}
                    >
                        {value ? format(value, "yyyy-MM-dd") : <span>{placeholder}</span>}
                        {showClear ? (
                            <>
                                <CalendarIcon className="size-4 group-hover/clear:hidden" />
                                <span
                                    role="button"
                                    tabIndex={-1}
                                    aria-label="Clear date"
                                    className="hidden rounded-sm p-0.5 text-muted-foreground hover:text-foreground group-hover/clear:inline-flex"
                                    onPointerDown={e => {
                                        e.preventDefault();
                                        e.stopPropagation();
                                    }}
                                    onClick={e => {
                                        e.preventDefault();
                                        e.stopPropagation();
                                        onChange?.(undefined);
                                    }}
                                >
                                    <X className="size-3.5" />
                                </span>
                            </>
                        ) : (
                            <CalendarIcon className="size-4" />
                        )}
                    </Button>
                </PopoverTrigger>
            </div>
            <PopoverContent
                className="w-auto p-0"
                align="start"
            >
                <Calendar
                    mode="single"
                    selected={value ?? undefined}
                    onSelect={onChange}
                />
            </PopoverContent>
        </Popover>
    );
}
