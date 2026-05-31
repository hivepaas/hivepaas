import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditAppScheduledJobDialog } from "~/projects/dialogs/create-or-edit-app-scheduled-job";
import type { AppScheduledJob } from "~/projects/domain";

function View({ projectId, appId, scheduledJob }: Props) {
    const { actions } = useCreateOrEditAppScheduledJobDialog();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                actions.openEdit(projectId, appId, scheduledJob.id);
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
