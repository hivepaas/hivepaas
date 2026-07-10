import { useMemo, useState } from "react";

import { Field, FieldError } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { Link, useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectRegistryAuthQueries } from "~/projects/data/queries";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { Combobox, InfoBlock } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";

import {
    type AppConfigDeploymentSettingsFormSchemaInput,
    type AppConfigDeploymentSettingsFormSchemaOutput,
} from "../../schemas";

export function PushToRegistrySelect({ readOnly = false }: Props) {
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
                label: `${auth.kind} ${auth.name}`,
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
                        if (readOnly) {
                            return;
                        }

                        pushToRegistry.onChange(option ?? undefined);
                    }}
                    onSearch={setSearchQuery}
                    placeholder="Select registry to push image to"
                    searchable
                    allowClear
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
                    <Link
                        to="#"
                        className="text-blue-500"
                    >
                        Configure Registry Credentials
                    </Link>
                </div>
            </Field>
        </InfoBlock>
    );
}

type Props = {
    readOnly?: boolean;
};
