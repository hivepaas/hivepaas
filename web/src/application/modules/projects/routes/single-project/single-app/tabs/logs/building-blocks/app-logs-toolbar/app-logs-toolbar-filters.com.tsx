import { DateTimePicker } from "@/components/ui/date-time-picker";
import { InputNumber } from "@/components/ui/input-number";

import { DurationPicker } from "../duration-picker";
import { LOG_FILTER_FIELD_WIDTH } from "../log-filters.constants";

/**
 * AppLogsToolbarFilters is the time window both log views are read through.
 *
 * Since and Duration compose the same way on the server for stored logs as for
 * the live stream: duration alone means the last N up to now, Since alone means
 * from then until now, and together they mean a fixed window starting at Since.
 * Lines means something slightly different on each side - a tail for the stream,
 * a page size for stored logs - but it is the same promise either way, and it
 * always applies: no combination of these controls asks for an unbounded read.
 */
export function AppLogsToolbarFilters({
    lines,
    since,
    duration,
    durationOptions,
    sinceFromDate,
    onLinesChange,
    onSinceChange,
    onDurationChange,
    onRequestRefresh,
}: AppLogsToolbarFiltersProps) {
    return (
        <div className="flex min-w-0 flex-wrap items-center gap-1.5 sm:gap-2">
            <div className="flex items-center gap-1 sm:gap-2">
                <span className="text-xs sm:text-sm font-medium text-foreground">Lines</span>
                <InputNumber
                    value={lines}
                    min={1}
                    showControls={false}
                    useGrouping={false}
                    className="w-[56px] sm:w-[76px]"
                    classNameInput="h-8 sm:h-9"
                    classNameInner="text-xs sm:text-sm px-2"
                    onValueChange={value => {
                        onLinesChange(value && value > 0 ? value : undefined);
                    }}
                    onKeyDown={event => {
                        if (event.key === "Enter") {
                            onRequestRefresh();
                        }
                    }}
                />
            </div>
            <DateTimePicker
                value={since}
                onChange={value => {
                    onSinceChange(value);
                    onRequestRefresh();
                }}
                placeholder="Since"
                showClearButton
                granularity="second"
                fromDate={sinceFromDate}
                containerClassName={LOG_FILTER_FIELD_WIDTH}
                className="h-8 sm:h-9 text-xs sm:text-sm px-2 sm:px-3"
                toDate={new Date()}
            />
            <DurationPicker
                value={duration}
                options={durationOptions}
                onChange={value => {
                    onDurationChange(value);
                    onRequestRefresh();
                }}
            />
        </div>
    );
}

interface AppLogsToolbarFiltersProps {
    lines: number | undefined;
    since: Date | undefined;
    duration: string | undefined;
    /** Shortcuts the duration popover lists. Defaults to the stream's set. */
    durationOptions?: readonly string[];
    /** Earliest instant Since may point at, when retention bounds it. */
    sinceFromDate?: Date;
    onLinesChange: (value: number | undefined) => void;
    onSinceChange: (value: Date | undefined) => void;
    onDurationChange: (value: string | undefined) => void;
    onRequestRefresh: () => void;
}
