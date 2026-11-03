import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import type { AppHealthCheck } from "~/projects/domain";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

function View({ projectId, env, appId, healthCheck }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(
                    ROUTE.projects.single.apps.single.configuration.periodicJobs.edit.$route(
                        projectId,
                        env,
                        appId,
                        healthCheck.id,
                    ),
                );
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit app health check</span>
        </Button>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    healthCheck: AppHealthCheck;
}

export const EditCell = memo(View);
