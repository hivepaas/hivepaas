import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { BasicAuthTableScope } from "../../basic-auth-table.types";

function View({ scope, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(getBasicAuthEditRoute(scope, id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit basic auth</span>
        </Button>
    );
}

function getBasicAuthEditRoute(scope: BasicAuthTableScope, id: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.basicAuth.edit.$route(scope.projectId, id);
    }

    return ROUTE.settings.basicAuth.edit.$route(id);
}

interface Props {
    scope: BasicAuthTableScope;
    id: string;
}

export const BasicAuthEditCell = memo(View);
