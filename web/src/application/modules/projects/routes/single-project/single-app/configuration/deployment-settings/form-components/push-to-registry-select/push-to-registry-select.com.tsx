import { useMemo, useState } from "react";

import { Field, FieldError } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { Link, useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectRegistryAuthQueries } from "~/projects/data/queries";

import { Combobox, InfoBlock } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../../schemas";

export function PushToRegistrySelect() {
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
        field: pushToRegistry,
        fieldState: { invalid: isPushToRegistryInvalid, error: pushToRegistryError },
    } = useController({ control, name: "repoSource.pushToRegistry" });

    const comboboxOptions = useMemo(() => {
        return registryAuths.map(auth => {
            return {
                value: { id: auth.id, name: auth.name },
                label: auth.name,
            };
        });
    }, [registryAuths]);

    return (
        <InfoBlock title="Registry To Push Image To">
            <Field>
                <Combobox
                    options={comboboxOptions}
                    value={pushToRegistry.value?.id ?? null}
                    onChange={(_, option) => {
                        pushToRegistry.onChange(option ?? null);
                    }}
                    onSearch={setSearchQuery}
                    placeholder="Select registry to push image to"
                    searchable
                    closeOnSelect
                    emptyText="No registry to push image to available"
                    className="max-w-[400px]"
                    valueKey="id"
                    aria-invalid={isPushToRegistryInvalid}
                    loading={isFetching}
                    onRefresh={() => void refetch()}
                    isRefreshing={isRefetching}
                    splitLabelBadge
                />
                <FieldError errors={[pushToRegistryError]} />
                <div className="text-xs">
                    <p>
                        Need to add a new registry to push image to?{" "}
                        <Link
                            to="#"
                            className="text-blue-500"
                        >
                            Click here
                        </Link>
                    </p>
                </div>
            </Field>
        </InfoBlock>
    );
}
