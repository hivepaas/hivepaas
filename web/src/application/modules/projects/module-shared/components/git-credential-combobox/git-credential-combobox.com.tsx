import { useMemo, useState } from "react";

import { ProjectGitCredentialsQueries } from "~/projects/data/queries";

import { AppLink, Combobox, type ComboboxOption } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { ESettingType } from "@application/shared/enums";

import { Badge, type BadgeTone } from "@/components/ui/badge";

export type GitCredentialOption = { id: string; name: string; type?: string; kind?: string };

function getGitCredentialBadge(cred?: { type?: string; kind?: string }): { label: string; tone: BadgeTone } {
    const type = cred?.type ?? "";
    const defaultKind = type === ESettingType.GithubApp || type === "github-app" ? "github" : "git";
    const rawKind = (cred?.kind ?? defaultKind).toLowerCase();
    const cleanKind = rawKind.replace(/-(app|token|ssh-key)$/, "");

    if (type === ESettingType.GithubApp || type === "github-app") {
        return {
            label: "github-app",
            tone: "purple",
        };
    }
    if (type === ESettingType.SSHKey || type === "ssh-key") {
        return {
            label: `${cleanKind}-ssh-key`,
            tone: "emerald",
        };
    }
    // Token / Access Token
    return {
        label: `${cleanKind}-token`,
        tone: "blue",
    };
}

/**
 * A pick among the git credentials an env of a project can use.
 */
export function GitCredentialCombobox({
    projectId,
    env,
    value,
    onChange,
    readOnly = false,
    invalid = false,
    className,
}: GitCredentialComboboxProps) {
    const [searchQuery, setSearchQuery] = useState("");

    const {
        data: { data: credentials } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = ProjectGitCredentialsQueries.useFindManyPaginated({
        projectID: projectId,
        env,
        search: searchQuery,
    });

    const comboboxOptions = useMemo(() => {
        return credentials.map(cred => {
            const badge = getGitCredentialBadge(cred);
            return {
                value: { id: cred.id, name: cred.name, type: cred.type, kind: cred.kind },
                label: `${badge.label} ${cred.name}`,
            };
        });
    }, [credentials]);

    const renderCredentialOption = (option: ComboboxOption<GitCredentialOption>) => {
        const cred = option.value;
        const matchedCred = credentials.find(c => c.id === cred.id);
        const badge = getGitCredentialBadge(matchedCred ?? cred);
        return (
            <span className="flex min-w-0 max-w-full items-center gap-2 text-left">
                <Badge
                    tone={badge.tone}
                    className="max-w-none shrink-0 rounded-md px-1.5 text-xs font-medium leading-none"
                >
                    {badge.label}
                </Badge>
                <span className="min-w-0 flex-1 truncate font-normal">{cred.name || matchedCred?.name}</span>
            </span>
        );
    };

    return (
        <Combobox
            options={comboboxOptions}
            value={value?.id ?? null}
            onChange={(_, option) => {
                if (readOnly) {
                    return;
                }

                onChange(option ?? null);
            }}
            onSearch={setSearchQuery}
            placeholder="Select git credentials"
            searchable
            closeOnSelect
            emptyText="No git credentials available"
            className={className}
            valueKey="id"
            aria-invalid={invalid}
            loading={isFetching}
            onRefresh={() => void refetch()}
            isRefreshing={isRefetching}
            renderOption={renderCredentialOption}
            renderSelectedOption={renderCredentialOption}
            disabled={readOnly}
        />
    );
}

/**
 * Where a project's git credentials are configured.
 */
export function GitCredentialLinks({ projectId }: { projectId: string }) {
    return (
        <div className="text-xs">
            Configure{" "}
            <AppLink.Basic
                to={ROUTE.projects.single.providerConfiguration.githubApps.$route(projectId)}
                className="text-link"
                target="_blank"
                rel="noopener noreferrer"
            >
                Github Apps
            </AppLink.Basic>
            ,{" "}
            <AppLink.Basic
                to={ROUTE.projects.single.providerConfiguration.accessTokens.$route(projectId)}
                className="text-link"
                target="_blank"
                rel="noopener noreferrer"
            >
                Access Tokens
            </AppLink.Basic>
            ,{" "}
            <AppLink.Basic
                to={ROUTE.projects.single.providerConfiguration.sshKeys.$route(projectId)}
                className="text-link"
                target="_blank"
                rel="noopener noreferrer"
            >
                SSH Keys
            </AppLink.Basic>
        </div>
    );
}

export interface GitCredentialComboboxProps {
    projectId: string;
    env: string;
    value: GitCredentialOption | null;
    onChange: (value: GitCredentialOption | null) => void;
    readOnly?: boolean;
    invalid?: boolean;
    className?: string;
}
