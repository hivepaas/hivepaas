import { useMemo, useState } from "react";

import { Field, FieldError, FieldGroup } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { Link, useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectGitCredentialsQueries } from "~/projects/data/queries";

import { Combobox, InfoBlock } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../../schemas";

export function GitCredentialSelect() {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

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
        search: searchQuery,
    });

    const {
        field: credentialsField,
        fieldState: { invalid: isCredentialsInvalid, error: credentialsError },
    } = useController({ control, name: "repoSource.credentials" });

    const comboboxOptions = useMemo(() => {
        return credentials.map(cred => {
            return {
                value: { id: cred.id, name: cred.name },
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
                            credentialsField.onChange(option ?? null);
                        }}
                        onSearch={setSearchQuery}
                        placeholder="Select git credential"
                        searchable
                        allowClear
                        closeOnSelect
                        emptyText="No git credentials available"
                        className="max-w-[400px]"
                        valueKey="id"
                        aria-invalid={isCredentialsInvalid}
                        loading={isFetching}
                        onRefresh={() => void refetch()}
                        isRefreshing={isRefetching}
                        splitLabelBadge
                    />
                    <FieldError errors={[credentialsError]} />
                    <div className="text-xs">
                        <p>
                            Need to add a new git credential?{" "}
                            <Link
                                to="#"
                                className="text-blue-500"
                            >
                                Click here
                            </Link>
                        </p>
                    </div>
                </Field>
            </FieldGroup>
        </InfoBlock>
    );
}
