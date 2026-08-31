import { useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { Field, FieldError, FieldGroup } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectGitCredentialsQueries } from "~/projects/data/queries";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { AppLink, Combobox, type ComboboxOption, InfoBlock } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { ESettingType } from "@application/shared/enums";

import { Badge } from "@/components/ui/badge";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../../schemas";

function getGitCredentialBadge(cred?: { type?: string; kind?: string }) {
    const type = cred?.type ?? "";
    const defaultKind = type === ESettingType.GithubApp || type === "github-app" ? "github" : "git";
    const rawKind = (cred?.kind ?? defaultKind).toLowerCase();
    const cleanKind = rawKind.replace(/-(app|token|ssh-key)$/, "");

    if (type === ESettingType.GithubApp || type === "github-app") {
        return {
            label: "github-app",
            className: "bg-purple-600 text-white dark:bg-purple-600",
        };
    }
    if (type === ESettingType.SSHKey || type === "ssh-key") {
        return {
            label: `${cleanKind}-ssh-key`,
            className: "bg-emerald-600 text-white dark:bg-emerald-600",
        };
    }
    // Token / Access Token
    return {
        label: `${cleanKind}-token`,
        className: "bg-blue-600 text-white dark:bg-blue-600",
    };
}

export function GitCredentialSelect({ readOnly = false }: Props) {
    const { id: projectId, env } = useParams<{ id: string; env: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");

    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

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

    const {
        field: credentialsField,
        fieldState: { invalid: isCredentialsInvalid, error: credentialsError },
    } = useController({ control, name: "repoSource.credentials" });

    const comboboxOptions = useMemo(() => {
        return credentials.map(cred => {
            const badge = getGitCredentialBadge(cred);
            return {
                value: { id: cred.id, name: cred.name, type: cred.type, kind: cred.kind },
                label: `${badge.label} ${cred.name}`,
            };
        });
    }, [credentials]);

    const renderCredentialOption = (
        option: ComboboxOption<{ id: string; name: string; type?: string; kind?: string }>,
    ) => {
        const cred = option.value;
        const matchedCred = credentials.find(c => c.id === cred.id);
        const badge = getGitCredentialBadge(matchedCred ?? cred);
        return (
            <span className="flex min-w-0 max-w-full items-center gap-2 text-left">
                <Badge
                    className={cn(
                        "max-w-none shrink-0 rounded-md px-1.5 text-xs font-medium leading-none border-transparent",
                        badge.className,
                    )}
                >
                    {badge.label}
                </Badge>
                <span className="min-w-0 flex-1 truncate font-normal">{cred.name || matchedCred?.name}</span>
            </span>
        );
    };

    return (
        <InfoBlock
            titleWidth={220}
            title="Git Credentials"
        >
            <FieldGroup>
                <Field>
                    <Combobox
                        options={comboboxOptions}
                        value={credentialsField.value?.id ?? null}
                        onChange={(_, option) => {
                            if (readOnly) {
                                return;
                            }

                            credentialsField.onChange(option ?? null);
                        }}
                        onSearch={setSearchQuery}
                        placeholder="Select git credentials"
                        searchable
                        closeOnSelect
                        emptyText="No git credentials available"
                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                        valueKey="id"
                        aria-invalid={isCredentialsInvalid}
                        loading={isFetching}
                        onRefresh={() => void refetch()}
                        isRefreshing={isRefetching}
                        renderOption={renderCredentialOption}
                        renderSelectedOption={renderCredentialOption}
                        disabled={readOnly}
                    />
                    <FieldError errors={[credentialsError]} />
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
                </Field>
            </FieldGroup>
        </InfoBlock>
    );
}

type Props = {
    readOnly?: boolean;
};
