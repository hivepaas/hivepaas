import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { SslProviderTableScope } from "../../ssl-provider-table.types";

function View({ scope, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(getSslProviderEditRoute(scope, id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit SSL provider</span>
        </Button>
    );
}

function getSslProviderEditRoute(scope: SslProviderTableScope, id: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.sslProviders.edit.$route(scope.projectId, id);
    }

    return ROUTE.settings.sslProviders.edit.$route(id);
}

interface Props {
    scope: SslProviderTableScope;
    id: string;
}

export const SslProviderEditCell = memo(View);
