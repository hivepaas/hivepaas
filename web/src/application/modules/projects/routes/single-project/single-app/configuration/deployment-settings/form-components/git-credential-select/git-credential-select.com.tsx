import { useMemo, useState } from "react";

import { Field, FieldError, FieldGroup } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectGitCredentialsQueries } from "~/projects/data/queries";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { AppLink, Combobox, InfoBlock } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../../schemas";

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
            return {
                value: { id: cred.id, name: cred.name, type: cred.type },
                label: `${cred.type} ${cred.name}`,
            };
        });
    }, [credentials]);

    return (
        <InfoBlock title="Git Credentials">
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
                        allowClear
                        closeOnSelect
                        emptyText="No git credentials available"
                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                        valueKey="id"
                        aria-invalid={isCredentialsInvalid}
                        loading={isFetching}
                        onRefresh={() => void refetch()}
                        isRefreshing={isRefetching}
                        splitLabelBadge
                        disabled={readOnly}
                    />
                    <FieldError errors={[credentialsError]} />
                    <div className="text-xs">
                        Configure{" "}
                        <AppLink.Basic
                            to={ROUTE.projects.single.sources.githubApps.$route(projectId)}
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
