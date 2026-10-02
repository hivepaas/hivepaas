import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SectionHeader } from "~/system-settings/module-shared";

import { NextRunsField } from "@application/shared/components";

import { ScheduleFields } from "./schedule-fields.com";

export function GeneralFields({ nextRuns }: Props) {
    return (
        <>
            <SectionHeader>General</SectionHeader>
            <div className="flex flex-col gap-6 px-3">
                <div className={cn(dashedBorderBox)}>
                    <span className="font-semibold text-orange-500">Note:</span>{" "}
                    <span>
                        An Amazon ECR credential signs in with a token that expires after 12 hours. Each run gets a new
                        token for every Amazon ECR credential and hands it to the services that pull with it, so that
                        Swarm can start them on any node. Running containers are not restarted. With no Amazon ECR
                        credential, a run does nothing.
                    </span>
                </div>

                <ScheduleFields />
                <NextRunsField
                    nextRuns={nextRuns}
                    titleWidth={220}
                />
            </div>
        </>
    );
}

interface Props {
    nextRuns: Date[];
}
