import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { SSHKeyTableScope } from "../ssh-key-table.types";

function View({ scope, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(getSSHKeyEditRoute(scope, id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit SSH key</span>
        </Button>
    );
}

function getSSHKeyEditRoute(scope: SSHKeyTableScope, id: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.sshKeys.edit.$route(scope.projectId, id);
    }

    return ROUTE.settings.sshKeys.edit.$route(id);
}

interface Props {
    scope: SSHKeyTableScope;
    id: string;
}

export const SSHKeyEditCell = memo(View);
