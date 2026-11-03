import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { RegistryAuthTableScope } from "../../registry-auth-table.types";

function View({ scope, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(getRegistryAuthEditRoute(scope, id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit registry auth</span>
        </Button>
    );
}

function getRegistryAuthEditRoute(scope: RegistryAuthTableScope, id: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.registryAuth.edit.$route(scope.projectId, id);
    }

    return ROUTE.settings.registryAuth.edit.$route(id);
}

interface Props {
    scope: RegistryAuthTableScope;
    id: string;
}

export const RegistryAuthEditCell = memo(View);
