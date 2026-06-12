import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";

import { NextRunsField } from "./next-runs-field.com";
import { ScheduleFields } from "./schedule-fields.com";
import { SchedulingModeField } from "./scheduling-mode-field.com";
import { SectionHeader } from "./section-header.com";

export function GeneralFields({ nextRuns }: Props) {
    return (
        <>
            <SectionHeader>General</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <div className={cn(dashedBorderBox, "text-center text-sm leading-6")}>
                    <span className="text-orange-500">Note:</span>{" "}
                    <span>
                        We encourage you to run this task during low server load periods (e.g., midnight). Additionally,
                        you should schedule system tasks at different times (e.g., system backup at 1 AM, followed by
                        certificate renewal at 2 AM).
                    </span>
                </div>

                <SchedulingModeField />
                <ScheduleFields />
                <NextRunsField nextRuns={nextRuns} />
            </div>
        </>
    );
}

interface Props {
    nextRuns: Date[];
}
