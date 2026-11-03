import { useEffect, useState } from "react";

import { cn } from "@/lib/utils";
import { ChevronDown, X } from "lucide-react";

import { Button, Input } from "@/components/ui";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

import { LOG_FILTER_FIELD_WIDTH } from "../log-filters.constants";

import { DEFAULT_DURATION_OPTIONS } from "./duration-picker.constants";

/**
 * DurationPicker edits a duration string such as `30m`, `7d` or `1d12h`.
 *
 * The value is never parsed here: it goes to the server as written, and the
 * server reads it with the same parser that reads the logging retention
 * setting. Listing shortcuts is all this component does with the units.
 */
export function DurationPicker({ value, options = DEFAULT_DURATION_OPTIONS, onChange }: DurationPickerProps) {
    const [open, setOpen] = useState(false);
    const [draft, setDraft] = useState(value ?? "");

    useEffect(() => {
        setDraft(value ?? "");
    }, [value]);

    function commitDuration(nextValue: string) {
        const normalizedValue = nextValue.trim();

        onChange(normalizedValue || undefined);
        setDraft(normalizedValue);
        setOpen(false);
    }

    return (
        <div className={cn("relative", LOG_FILTER_FIELD_WIDTH)}>
            <Popover
                open={open}
                onOpenChange={setOpen}
            >
                <PopoverTrigger asChild>
                    <Button
                        type="button"
                        variant="outline"
                        className={cn(
                            "h-8 sm:h-9 w-full justify-between px-2 sm:px-3 text-xs sm:text-sm font-normal",
                            !value && "text-muted-foreground",
                        )}
                    >
                        <span className="truncate">{value ?? "Duration"}</span>
                        <ChevronDown className="size-3.5 sm:size-4 opacity-50 shrink-0" />
                    </Button>
                </PopoverTrigger>
                <PopoverContent
                    align="start"
                    className="w-[180px] p-2"
                >
                    <Input
                        value={draft}
                        placeholder="Duration"
                        className="mb-2 h-8 sm:h-9 text-xs sm:text-sm"
                        onChange={event => {
                            setDraft(event.target.value);
                        }}
                        onKeyDown={event => {
                            if (event.key === "Enter") {
                                commitDuration(draft);
                            }
                        }}
                    />
                    <div className="flex max-h-[260px] flex-col overflow-y-auto">
                        {options.map(option => (
                            <button
                                key={option}
                                type="button"
                                className="rounded-sm px-2 py-1.5 text-left text-xs sm:text-sm hover:bg-accent hover:text-accent-foreground"
                                onClick={() => {
                                    commitDuration(option);
                                }}
                            >
                                {option}
                            </button>
                        ))}
                    </div>
                </PopoverContent>
            </Popover>
            {value && (
                <button
                    type="button"
                    className="absolute right-7 top-1/2 flex size-5 -translate-y-1/2 items-center justify-center rounded-sm text-muted-foreground hover:text-foreground"
                    onClick={() => {
                        onChange(undefined);
                    }}
                >
                    <X className="size-3.5" />
                </button>
            )}
        </div>
    );
}

interface DurationPickerProps {
    value: string | undefined;
    /** Shortcuts listed under the text field. Defaults to the live stream's set. */
    options?: readonly string[];
    onChange: (value: string | undefined) => void;
}
