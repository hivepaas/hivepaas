import React, { useState } from "react";

import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { format } from "date-fns";
import { ChevronDown } from "lucide-react";

import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui";

import { InfoBlock } from "../info-block";

function View({ nextRuns, titleWidth }: Props) {
    const [open, setOpen] = useState(false);

    if (nextRuns.length === 0) {
        return null;
    }

    const firstRun = nextRuns[0]!;
    const restRuns = nextRuns.slice(1);
    const hasMore = restRuns.length > 0;

    return (
        <InfoBlock
            title="Next Runs"
            titleWidth={titleWidth}
        >
            <Collapsible
                open={open}
                onOpenChange={setOpen}
                className="max-w-[400px]"
            >
                <div
                    className={cn(
                        dashedBorderBox,
                        "flex min-h-10 flex-col items-center justify-center gap-1 py-1.5 sm:py-1.5 text-center text-sm leading-6",
                    )}
                >
                    <div className="relative flex w-full items-center justify-center">
                        <span className="text-orange-500">{format(firstRun, "yyyy-MM-dd HH:mm:ss")}</span>
                        {hasMore && (
                            <CollapsibleTrigger asChild>
                                <button
                                    type="button"
                                    aria-label={open ? "Collapse next runs" : "Expand next runs"}
                                    className="absolute right-2 top-1/2 flex -translate-y-1/2 items-center justify-center"
                                >
                                    <ChevronDown
                                        className={cn(
                                            "size-4 text-muted-foreground transition-transform duration-200",
                                            open && "rotate-180",
                                        )}
                                    />
                                </button>
                            </CollapsibleTrigger>
                        )}
                    </div>
                    {hasMore && (
                        <CollapsibleContent className="flex flex-col gap-1">
                            {restRuns.map(runAt => (
                                <span
                                    key={runAt.toISOString()}
                                    className="text-orange-500"
                                >
                                    {format(runAt, "yyyy-MM-dd HH:mm:ss")}
                                </span>
                            ))}
                        </CollapsibleContent>
                    )}
                </div>
            </Collapsible>
        </InfoBlock>
    );
}

interface Props {
    nextRuns: Date[];
    titleWidth?: number;
}

export const NextRunsField = React.memo(View);
