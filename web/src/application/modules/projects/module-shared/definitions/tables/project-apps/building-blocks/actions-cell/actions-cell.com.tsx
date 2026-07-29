import React from "react";

import { EyeIcon } from "lucide-react";

import { AppLink } from "@application/shared/components/navigation";
import { ROUTE } from "@application/shared/constants";

function View({ projectId, env, appId }: Props) {
    return (
        <div className="flex items-center justify-center gap-4">
            <AppLink.Modules
                className="flex items-center justify-center text-link hover:opacity-50 transition-opacity duration-200"
                to={ROUTE.projects.single.apps.single.configuration.general.$route(projectId, env, appId)}
            >
                <EyeIcon className="size-5" />
            </AppLink.Modules>
        </div>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
}

export const ActionsCell = React.memo(View);
