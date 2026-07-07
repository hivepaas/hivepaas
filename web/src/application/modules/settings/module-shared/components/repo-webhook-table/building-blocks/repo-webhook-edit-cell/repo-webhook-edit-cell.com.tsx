import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { RepoWebhookTableScope } from "../../repo-webhook-table.types";

function View({ scope, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(getRepoWebhookEditRoute(scope, id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit webhook</span>
        </Button>
    );
}

interface Props {
    scope: RepoWebhookTableScope;
    id: string;
}

export const RepoWebhookEditCell = memo(View);

function getRepoWebhookEditRoute(scope: RepoWebhookTableScope, id: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.sources.webhooks.edit.$route(scope.projectId, id);
    }

    return ROUTE.sources.webhooks.edit.$route(id);
}
