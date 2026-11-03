import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

function View({ projectId, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(
                    ROUTE.projects.single.providerConfiguration.commandTemplates.edit.$route(projectId, id),
                );
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit command template</span>
        </Button>
    );
}

interface Props {
    projectId: string;
    id: string;
}

export const ProjectCommandTemplateEditCell = memo(View);
