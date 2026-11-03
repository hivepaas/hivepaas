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
    "disableClear"?: boolean;
}

export function DatePicker({
    value,
    onChange,
    placeholder = "Pick a date",
    disabled,
    className,
    "aria-invalid": ariaInvalid,
    disableClear = false,
}: DatePickerProps) {
    const showClear = !disableClear && !disabled && Boolean(value);

    return (
        <Popover>
            <div className="group/clear relative">
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
                        <CalendarIcon
                            className={cn(
                                "size-4",
                                showClear && "group-hover/clear:invisible group-focus-within/clear:invisible",
                            )}
                        />
                    </Button>
                </PopoverTrigger>
                {/* Over the calendar's icon, while the field is pointed at or in focus: a
                    button of its own, beside the one that opens the calendar. */}
                {showClear && (
                    <button
                        type="button"
                        aria-label="Clear date"
                        className="absolute right-3 top-1/2 hidden -translate-y-1/2 rounded-sm p-0.5 text-muted-foreground hover:text-foreground group-focus-within/clear:inline-flex group-hover/clear:inline-flex focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                        onClick={() => {
                            onChange?.(undefined);
                        }}
                    >
                        <X className="size-3.5" />
                    </button>
                )}
            </div>
            <PopoverContent
                className="w-auto p-0"
                align="start"
            >
                <Calendar
                    mode="single"
                    selected={value ?? undefined}
                    onSelect={onChange}
                    captionLayout="dropdown"
                    startMonth={new Date(new Date().getFullYear() - 50, 0)}
                    endMonth={new Date(new Date().getFullYear() + 50, 11)}
                />
            </PopoverContent>
        </Popover>
    );
}
