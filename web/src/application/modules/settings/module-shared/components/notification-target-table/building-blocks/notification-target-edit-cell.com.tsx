import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { NotificationTargetTableScope } from "../notification-target-table.types";

function View({ scope, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(getNotificationTargetEditRoute(scope, id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit notification target</span>
        </Button>
    );
}

function getNotificationTargetEditRoute(scope: NotificationTargetTableScope, id: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.providerConfiguration.notificationTargets.edit.$route(scope.projectId, id);
    }

    return ROUTE.settings.notificationTargets.edit.$route(id);
}

interface Props {
    scope: NotificationTargetTableScope;
    id: string;
}

export const NotificationTargetEditCell = memo(View);
