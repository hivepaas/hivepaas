import React from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import type { AppSettingMount } from "~/projects/domain";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

function View({ projectId, env, appId, settingMount }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(
                    ROUTE.projects.single.apps.single.configuration.settingMounts.edit.$route(
                        projectId,
                        env,
                        appId,
                        settingMount.id,
                    ),
                );
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit app setting mount</span>
        </Button>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    settingMount: AppSettingMount;
}

export const EditCell = React.memo(View);
