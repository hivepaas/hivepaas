import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { CloudStorageTableScope } from "../cloud-storage-table.types";

function View({ scope, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(getCloudStorageEditRoute(scope, id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit cloud storage</span>
        </Button>
    );
}

function getCloudStorageEditRoute(scope: CloudStorageTableScope, id: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.cloudStorages.edit.$route(scope.projectId, id);
    }

    return ROUTE.settings.cloudStorages.edit.$route(id);
}

interface Props {
    scope: CloudStorageTableScope;
    id: string;
}

export const CloudStorageEditCell = memo(View);
