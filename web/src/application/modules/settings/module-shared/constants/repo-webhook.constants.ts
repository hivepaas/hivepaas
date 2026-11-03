import { ERepoWebhookKind } from "~/settings/module-shared/enums";

export interface RepoWebhookKindOption {
    value: string;
    label: string;
}

export const REPO_WEBHOOK_KIND_LABELS: Record<string, string> = {
    [ERepoWebhookKind.Github]: "Github",
    [ERepoWebhookKind.Gitlab]: "Gitlab",
    [ERepoWebhookKind.Gitea]: "Gitea",
    [ERepoWebhookKind.Bitbucket]: "Bitbucket",
    [ERepoWebhookKind.Gogs]: "Gogs",
    "": "Any Provider",
};

export const REPO_WEBHOOK_KIND_OPTIONS: readonly RepoWebhookKindOption[] = [
    {
        value: ERepoWebhookKind.Github,
        label: "Github",
    },
    {
        value: ERepoWebhookKind.Gitlab,
        label: "Gitlab",
    },
    {
        value: ERepoWebhookKind.Gitea,
        label: "Gitea",
    },
    {
        value: ERepoWebhookKind.Bitbucket,
        label: "Bitbucket",
    },
    {
        value: ERepoWebhookKind.Gogs,
        label: "Gogs",
    },
    {
        value: "",
        label: "Any Provider",
    },
] as const;

export function formatRepoWebhookKind(kind: string): string {
    return REPO_WEBHOOK_KIND_LABELS[kind] ?? kind;
}
