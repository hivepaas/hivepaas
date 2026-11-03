import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { AccessTokenTableScope } from "../access-token-table.types";

function View({ scope, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(getAccessTokenEditRoute(scope, id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit access token</span>
        </Button>
    );
}

function getAccessTokenEditRoute(scope: AccessTokenTableScope, id: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.accessTokens.edit.$route(scope.projectId, id);
    }

    return ROUTE.settings.accessTokens.edit.$route(id);
}

interface Props {
    scope: AccessTokenTableScope;
    id: string;
}

export const AccessTokenEditCell = memo(View);
