import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { SslCertTableScope } from "../../ssl-cert-table.types";

function View({ scope, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(getSslCertEditRoute(scope, id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit SSL certificate</span>
        </Button>
    );
}

function getSslCertEditRoute(scope: SslCertTableScope, id: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.sslCertificates.edit.$route(scope.projectId, id);
    }

    return ROUTE.settings.sslCertificates.edit.$route(id);
}

interface Props {
    scope: SslCertTableScope;
    id: string;
}

export const SslCertEditCell = memo(View);
