import * as React from "react";

import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { CalendarIcon } from "lucide-react";

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
}

export function DatePicker({
    value,
    onChange,
    placeholder = "Pick a date",
    disabled,
    className,
    "aria-invalid": ariaInvalid,
}: DatePickerProps) {
    return (
        <Popover>
            <PopoverTrigger asChild>
                <Button
                    variant="outline"
                    disabled={disabled}
                    aria-invalid={ariaInvalid}
                    className={cn(
                        "w-full justify-start text-left font-normal",
                        !value && "text-muted-foreground",
                        className,
                    )}
                >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {value ? format(value, "yyyy-MM-dd") : <span>{placeholder}</span>}
                </Button>
            </PopoverTrigger>
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
