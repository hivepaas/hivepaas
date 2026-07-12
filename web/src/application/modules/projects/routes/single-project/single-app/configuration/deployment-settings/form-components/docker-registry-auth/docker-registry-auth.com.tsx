import { useMemo, useState } from "react";

import { Field, FieldError, FieldGroup } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectRegistryAuthQueries } from "~/projects/data/queries";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { AppLink, Combobox, InfoBlock } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../../schemas";

export function DockerRegistryAuth({ readOnly = false }: Props) {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

    const { control } = useFormContext<
        AppConfigDeploymentSettingsFormSchemaInput,
        unknown,
        AppConfigDeploymentSettingsFormSchemaOutput
    >();

    const [searchQuery, setSearchQuery] = useState("");

    const {
        data: { data: registryAuths } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = ProjectRegistryAuthQueries.useFindManyPaginated({
        projectID: projectId,
        search: searchQuery,
    });

    const {
        field: registryAuth,
        fieldState: { invalid: isRegistryAuthInvalid, error: registryAuthError },
    } = useController({ control, name: "imageSource.registryAuth" });

    const comboboxOptions = useMemo(() => {
        return registryAuths.map(auth => {
            return {
                value: { id: auth.id, name: auth.name },
                label: `${auth.kind} ${auth.name}`,
            };
        });
    }, [registryAuths]);

    return (
        <InfoBlock title="Registry Credentials">
            <FieldGroup>
                <Field>
                    <Combobox
                        options={comboboxOptions}
                        value={registryAuth.value?.id ?? null}
                        onChange={(_, option) => {
                            if (readOnly) {
                                return;
                            }

                            registryAuth.onChange(option ?? undefined);
                        }}
                        onSearch={setSearchQuery}
                        placeholder="Select registry credentials"
                        searchable
                        closeOnSelect
                        emptyText="No registry credentials available"
                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                        valueKey="id"
                        aria-invalid={isRegistryAuthInvalid}
                        loading={isFetching}
                        onRefresh={() => void refetch()}
                        isRefreshing={isRefetching}
                        splitLabelBadge
                        disabled={readOnly}
                    />
                    <FieldError errors={[registryAuthError]} />
                    <div className="text-xs">
                        <AppLink.Basic
                            to={ROUTE.projects.single.providerConfiguration.registryAuth.$route(projectId)}
                            className="text-link"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            Configure Registry Credentials
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
