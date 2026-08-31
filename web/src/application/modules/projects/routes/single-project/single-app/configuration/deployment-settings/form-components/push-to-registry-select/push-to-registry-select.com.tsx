import { useMemo, useState } from "react";

import { Field, FieldError } from "@components/ui";
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

const NONE_REGISTRY_OPTION = {
    value: { id: "", name: "None" },
    label: "none None",
};

export function PushToRegistrySelect({ readOnly = false }: Props) {
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
        data: { data: registryAuths } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = ProjectRegistryAuthQueries.useFindManyPaginated({
        projectID: projectId,
        env,
        search: searchQuery,
    });

    const {
        field: pushToRegistry,
        fieldState: { invalid: isPushToRegistryInvalid, error: pushToRegistryError },
    } = useController({ control, name: "repoSource.pushToRegistry" });

    const comboboxOptions = useMemo(() => {
        const registryOptions = registryAuths.map(auth => {
            return {
                value: { id: auth.id, name: auth.name },
                label: `${auth.kind} ${auth.name}`,
            };
        });

        return [NONE_REGISTRY_OPTION, ...registryOptions];
    }, [registryAuths]);

    return (
        <InfoBlock
            titleWidth={220}
            title="Registry To Push Image To"
        >
            <Field>
                <Combobox
                    options={comboboxOptions}
                    value={pushToRegistry.value?.id ?? ""}
                    onChange={(_, option) => {
                        if (readOnly) {
                            return;
                        }

                        if (!option || option.id === "") {
                            pushToRegistry.onChange(undefined);
                            return;
                        }

                        pushToRegistry.onChange(option);
                    }}
                    onSearch={setSearchQuery}
                    placeholder="Select registry to push image to"
                    searchable
                    closeOnSelect
                    emptyText="No registry to push image to available"
                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                    valueKey="id"
                    aria-invalid={isPushToRegistryInvalid}
                    loading={isFetching}
                    onRefresh={() => void refetch()}
                    isRefreshing={isRefetching}
                    splitLabelBadge
                    disabled={readOnly}
                />
                <FieldError errors={[pushToRegistryError]} />
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
        </InfoBlock>
    );
}

type Props = {
    readOnly?: boolean;
};
