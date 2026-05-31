import React, { useState } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { format } from "date-fns";
import { ChevronDown } from "lucide-react";

import { InfoBlock } from "@application/shared/components";

import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui";

import { INFO_BLOCK_TITLE_WIDTH } from "./app-scheduled-job-form.constants";

function View({ nextRuns }: Props) {
    const [open, setOpen] = useState(false);

    if (nextRuns.length === 0) {
        return null;
    }

    return (
        <InfoBlock
            title="Next Runs"
            titleWidth={INFO_BLOCK_TITLE_WIDTH}
        >
            <Collapsible
                open={open}
                onOpenChange={setOpen}
                className="max-w-[400px]"
            >
                <div className={cn(dashedBorderBox, "text-center text-sm leading-6 p-2")}>
                    <CollapsibleTrigger asChild>
                        <button
                            type="button"
                            aria-label={open ? "Collapse next runs" : "Expand next runs"}
                            className="flex w-full items-center justify-end"
                        >
                            <ChevronDown
                                className={cn(
                                    "size-4 text-muted-foreground transition-transform duration-200",
                                    open && "rotate-180",
                                )}
                            />
                        </button>
                    </CollapsibleTrigger>
                    <CollapsibleContent>
                        <div className="flex flex-col gap-1 pt-3 text-sm">
                            {nextRuns.map(runAt => (
                                <span
                                    key={runAt.toISOString()}
                                    className="text-orange-500"
                                >
                                    {format(runAt, "yyyy-MM-dd HH:mm:ss")}
                                </span>
                            ))}
                        </div>
                    </CollapsibleContent>
                </div>
            </Collapsible>
        </InfoBlock>
    );
}

interface Props {
    nextRuns: Date[];
}

export const NextRunsField = React.memo(View);
