import { useMemo } from "react";

import { FieldError } from "@components/ui";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectGitCredentialsQueries } from "~/projects/data/queries";

import { Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useDebouncedSearch } from "@application/shared/hooks";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../../schemas";

type RepoOptionValue = {
    id: string;
    name: string;
    fullName: string;
    cloneURL: string;
};

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
        isFetching,
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

    const comboboxOptions = useMemo(() => {
        return repos.map(repo => ({
            value: {
                id: repo.id,
                name: repo.name,
                fullName: repo.fullName,
                cloneURL: repo.cloneURL,
            },
            label: repo.fullName !== "" ? repo.fullName : repo.name,
        }));
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
            <Combobox<RepoOptionValue>
                options={comboboxOptions}
                value={repoUrl.value || null}
                onChange={(url, _option) => {
                    repoUrl.onChange(url ?? "");
                }}
                onSearch={setSearch}
                placeholder={hasCredential ? "Select repository" : "Select a git credential first"}
                searchable
                closeOnSelect
                emptyText={hasCredential ? "No repositories available" : "Select a git credential first"}
                className="max-w-[400px]"
                valueKey="cloneURL"
                aria-invalid={isRepoUrlInvalid}
                loading={isFetching}
                disabled={!hasCredential}
                onRefresh={hasCredential ? () => void refetch() : undefined}
                isRefreshing={isRefetching}
                splitLabelBadge
            />
            <FieldError errors={[repoUrlError]} />
        </InfoBlock>
    );
}
