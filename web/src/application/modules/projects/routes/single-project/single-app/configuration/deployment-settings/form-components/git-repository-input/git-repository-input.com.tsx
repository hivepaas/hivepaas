import { useMemo } from "react";

import { FieldError } from "@components/ui";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectGitCredentialsQueries } from "~/projects/data/queries";

import { EditableCombobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useDebouncedSearch } from "@application/shared/hooks";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../../schemas";

export function GitRepositoryInput() {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const [debouncedSearch, setSearch] = useDebouncedSearch(250, "");

    const credentials = useWatch({ control, name: "repoSource.credentials" });
    const credentialId = credentials?.id;

    const {
        data: { data: repos } = DEFAULT_PAGINATED_DATA,
        refetch,
        isRefetching,
    } = ProjectGitCredentialsQueries.useFindManyRepos(
        {
            projectID: projectId,
            itemID: credentialId ?? "",
            search: debouncedSearch,
        },
        { enabled: Boolean(credentialId) },
    );

    const {
        field: repoUrl,
        fieldState: { invalid: isRepoUrlInvalid, error: repoUrlError },
    } = useController({ control, name: "repoSource.repoUrl" });

    const repoCloneUrls = useMemo(() => {
        const urls = repos.map(repo => repo.cloneURL).filter(Boolean);
        return [...new Set(urls)];
    }, [repos]);

    const hasCredential = Boolean(credentialId);

    return (
        <InfoBlock
            title={
                <LabelWithInfo
                    label="Git Repository"
                    isRequired
                />
            }
        >
            <EditableCombobox
                options={repoCloneUrls}
                value={repoUrl.value}
                onChange={v => {
                    repoUrl.onChange(v);
                }}
                placeholder={
                    hasCredential
                        ? "Search or paste repository URL"
                        : "Paste repository URL or select a credential for suggestions"
                }
                emptyText={
                    hasCredential ? "No repositories available" : "Select a credential for repository suggestions"
                }
                className="max-w-[400px]"
                aria-invalid={isRepoUrlInvalid}
                onRefresh={hasCredential ? () => void refetch() : undefined}
                isRefreshing={isRefetching}
            />
            <FieldError errors={[repoUrlError]} />
        </InfoBlock>
    );
}
