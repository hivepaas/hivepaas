import React from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import type { AppSecret } from "~/projects/domain";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

function View({ projectId, appId, secret }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(
                    ROUTE.projects.single.apps.single.configuration.secrets.edit.$route(projectId, appId, secret.id),
                );
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit app secret</span>
        </Button>
    );
}

interface Props {
    projectId: string;
    appId: string;
    secret: AppSecret;
}

export const EditCell = React.memo(View);
