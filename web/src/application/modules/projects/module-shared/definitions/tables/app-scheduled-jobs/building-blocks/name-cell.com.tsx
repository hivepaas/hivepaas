import { memo } from "react";

import { Badge } from "@components/ui/badge";
import type { AppScheduledJob } from "~/projects/domain";
import { formatJobTrigger } from "~/projects/module-shared/components/job-triggers-field/job-triggers.helpers";
import { EAppScheduledJobType } from "~/projects/module-shared/enums";

/** A job's name; a sequence is tagged, with its step count. */
function View({ job }: Props) {
    const steps = job.sequence?.steps.length ?? 0;

    return (
        <div className="flex flex-wrap items-center gap-2">
            <span>{job.name}</span>
            {job.jobType === EAppScheduledJobType.JobSequence && (
                <Badge variant="secondary">
                    Sequence · {steps} {steps === 1 ? "step" : "steps"}
                </Badge>
            )}
            {job.triggers.map(trigger => (
                <Badge
                    key={`${trigger.event}-${trigger.apps.map(app => app.id).join(",")}`}
                    variant="outline"
                    className="font-mono text-[11px]"
                    title={trigger.apps.length > 0 ? `of ${trigger.apps.map(app => app.name).join(", ")}` : undefined}
                >
                    {formatJobTrigger(trigger)}
                </Badge>
            ))}
        </div>
    );
}

interface Props {
    job: AppScheduledJob;
}

export const ScheduledJobNameCell = memo(View);
