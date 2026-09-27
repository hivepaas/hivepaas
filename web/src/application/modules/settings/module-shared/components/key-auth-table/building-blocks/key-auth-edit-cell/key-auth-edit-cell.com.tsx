import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { KeyAuthTableScope } from "../../key-auth-table.types";

function View({ scope, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(getKeyAuthEditRoute(scope, id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit key auth</span>
        </Button>
    );
}

function getKeyAuthEditRoute(scope: KeyAuthTableScope, id: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.keyAuth.edit.$route(scope.projectId, id);
    }

    return ROUTE.settings.keyAuth.edit.$route(id);
}

interface Props {
    scope: KeyAuthTableScope;
    id: string;
}

export const KeyAuthEditCell = memo(View);
