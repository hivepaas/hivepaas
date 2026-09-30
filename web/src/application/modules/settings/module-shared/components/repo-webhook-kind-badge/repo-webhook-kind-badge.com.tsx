import { memo } from "react";

import { Badge, type BadgeTone } from "@components/ui/badge";
import { formatRepoWebhookKind } from "~/settings/module-shared/constants";
import { ERepoWebhookKind } from "~/settings/module-shared/enums";

const KIND_TONES: Partial<Record<string, BadgeTone>> = {
    [ERepoWebhookKind.Github]: "sky",
    [ERepoWebhookKind.Gitlab]: "indigo",
    [ERepoWebhookKind.Gitea]: "fuchsia",
    [ERepoWebhookKind.Gogs]: "yellow",
    [ERepoWebhookKind.Bitbucket]: "purple",
};

function View({ kind }: Props) {
    const isKnownKind = Object.values(ERepoWebhookKind).includes(kind as ERepoWebhookKind);
    const label = formatRepoWebhookKind(kind);

    return <Badge tone={(isKnownKind ? KIND_TONES[kind] : undefined) ?? "purple"}>{label || "-"}</Badge>;
}

interface Props {
    kind: string;
}

export const RepoWebhookKindBadge = memo(View);
