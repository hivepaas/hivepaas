import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditAppHealthCheckDialog } from "~/projects/dialogs/create-or-edit-app-health-check";
import type { AppHealthCheck } from "~/projects/domain";

function View({ projectId, appId, healthCheck }: Props) {
    const { actions } = useCreateOrEditAppHealthCheckDialog();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                actions.openEdit(projectId, appId, healthCheck);
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit app health check</span>
        </Button>
    );
}

interface Props {
    projectId: string;
    appId: string;
    healthCheck: AppHealthCheck;
}

export const EditCell = memo(View);
