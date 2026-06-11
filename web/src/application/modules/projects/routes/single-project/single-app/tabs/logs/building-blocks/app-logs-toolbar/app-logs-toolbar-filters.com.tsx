import { useEffect, useState } from "react";

import { cn } from "@/lib/utils";
import { ChevronDown, X } from "lucide-react";

import { Button, Input } from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { InputNumber } from "@/components/ui/input-number";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

const DURATION_OPTIONS = ["5m", "15m", "30m", "1h", "2h", "4h", "1d", "2d", "4d"] as const;

export function AppLogsToolbarFilters({
    lines,
    since,
    duration,
    isLinesHidden,
    onLinesChange,
    onSinceChange,
    onDurationChange,
}: AppLogsToolbarFiltersProps) {
    return (
        <div className="flex min-w-0 flex-wrap items-center gap-2">
            {!isLinesHidden && (
                <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-foreground">Lines</span>
                    <InputNumber
                        value={lines}
                        min={1}
                        showControls={false}
                        useGrouping={false}
                        className="w-[76px]"
                        classNameInput="h-9"
                        onValueChange={value => {
                            onLinesChange(value && value > 0 ? value : undefined);
                        }}
                    />
                </div>
            )}
            <DateTimePicker
                value={since}
                onChange={onSinceChange}
                placeholder="Since"
                showClearButton
                granularity="second"
                containerClassName="w-[160px]"
                className="h-9"
                toDate={new Date()}
            />
            <DurationPicker
                value={duration}
                onChange={onDurationChange}
            />
        </div>
    );
}

function DurationPicker({ value, onChange }: DurationPickerProps) {
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
        <div className="relative w-[132px]">
            <Popover
                open={open}
                onOpenChange={setOpen}
            >
                <PopoverTrigger asChild>
                    <Button
                        type="button"
                        variant="outline"
                        className={cn("h-9 w-full justify-between px-3 font-normal", !value && "text-muted-foreground")}
                    >
                        <span className="truncate">{value ?? "Duration"}</span>
                        <ChevronDown className="size-4 opacity-50" />
                    </Button>
                </PopoverTrigger>
                <PopoverContent
                    align="start"
                    className="w-[180px] p-2"
                >
                    <Input
                        value={draft}
                        placeholder="Duration"
                        className="mb-2 h-9"
                        onChange={event => {
                            setDraft(event.target.value);
                        }}
                        onKeyDown={event => {
                            if (event.key === "Enter") {
                                commitDuration(draft);
                            }
                        }}
                    />
                    <div className="flex flex-col">
                        {DURATION_OPTIONS.map(option => (
                            <button
                                key={option}
                                type="button"
                                className="rounded-sm px-2 py-1.5 text-left text-sm hover:bg-accent hover:text-accent-foreground"
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

interface AppLogsToolbarFiltersProps {
    lines: number | undefined;
    since: Date | undefined;
    duration: string | undefined;
    isLinesHidden: boolean;
    onLinesChange: (value: number | undefined) => void;
    onSinceChange: (value: Date | undefined) => void;
    onDurationChange: (value: string | undefined) => void;
}

interface DurationPickerProps {
    value: string | undefined;
    onChange: (value: string | undefined) => void;
}
