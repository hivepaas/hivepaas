import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { ImPlatformTableScope } from "../../im-platform-table.types";

function View({ scope, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(getImPlatformEditRoute(scope, id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit IM platform</span>
        </Button>
    );
}

function getImPlatformEditRoute(scope: ImPlatformTableScope, id: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.imPlatforms.edit.$route(scope.projectId, id);
    }

    return ROUTE.settings.imPlatforms.edit.$route(id);
}

interface Props {
    scope: ImPlatformTableScope;
    id: string;
}

export const ImPlatformEditCell = memo(View);
