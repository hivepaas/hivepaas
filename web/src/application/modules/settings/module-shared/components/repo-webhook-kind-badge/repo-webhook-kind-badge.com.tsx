import { memo } from "react";

import { Badge } from "@components/ui/badge";
import { cn } from "@lib/utils";
import { ERepoWebhookKind } from "~/settings/module-shared/enums";

function View({ kind }: Props) {
    const isKnownKind = Object.values(ERepoWebhookKind).includes(kind as ERepoWebhookKind);

    return (
        <Badge
            className={cn(
                kind === ERepoWebhookKind.Github && "bg-sky-500 text-white",
                kind === ERepoWebhookKind.Gitlab && "bg-indigo-500 text-white",
                kind === ERepoWebhookKind.Gitea && "bg-fuchsia-500 text-white",
                kind === ERepoWebhookKind.Gogs && "bg-yellow-500 text-white",
                kind === ERepoWebhookKind.Bitbucket && "bg-purple-500 text-white",
                !isKnownKind && "bg-purple-500 text-white",
            )}
        >
            {kind || "-"}
        </Badge>
    );
}

interface Props {
    kind: string;
}

export const RepoWebhookKindBadge = memo(View);
