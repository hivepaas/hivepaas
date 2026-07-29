import React from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import type { AppConfigFile } from "~/projects/domain";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

function View({ projectId, env, appId, configFile }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(
                    ROUTE.projects.single.apps.single.configuration.configFiles.edit.$route(
                        projectId,
                        env,
                        appId,
                        configFile.id,
                    ),
                );
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit app config file</span>
        </Button>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    configFile: AppConfigFile;
}

export const EditCell = React.memo(View);
