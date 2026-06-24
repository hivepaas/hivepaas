import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import type { AppScheduledJob } from "~/projects/domain";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

function View({ projectId, appId, scheduledJob }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(
                    ROUTE.projects.single.apps.single.configuration.scheduledJobs.edit.$route(
                        projectId,
                        appId,
                        scheduledJob.id,
                    ),
                );
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit app scheduled job</span>
        </Button>
    );
}

interface Props {
    projectId: string;
    appId: string;
    scheduledJob: AppScheduledJob;
}

export const EditCell = memo(View);
