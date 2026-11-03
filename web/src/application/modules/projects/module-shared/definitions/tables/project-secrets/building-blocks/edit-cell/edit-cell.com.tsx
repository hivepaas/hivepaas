import React from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import type { ProjectSecret } from "~/projects/domain";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

function View({ projectId, secret }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(ROUTE.projects.single.providerConfiguration.secrets.edit.$route(projectId, secret.id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit project secret</span>
        </Button>
    );
}

interface Props {
    projectId: string;
    secret: ProjectSecret;
}

export const EditCell = React.memo(View);
