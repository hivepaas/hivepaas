import { memo } from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";

import { ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";

import type { GithubAppTableScope } from "../../github-app-table.types";

function View({ scope, id }: Props) {
    const { navigate } = useAppNavigate();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                navigate.modules(getGithubAppEditRoute(scope, id));
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit github app</span>
        </Button>
    );
}

interface Props {
    scope: GithubAppTableScope;
    id: string;
}

export const GithubAppEditCell = memo(View);

function getGithubAppEditRoute(scope: GithubAppTableScope, id: string) {
    if (scope.type === "project") {
        return ROUTE.projects.single.sources.githubApps.edit.$route(scope.projectId, id);
    }

    return ROUTE.sources.githubApps.edit.$route(id);
}
