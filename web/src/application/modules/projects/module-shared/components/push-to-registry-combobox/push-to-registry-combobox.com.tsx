import { useMemo, useState } from "react";

import { ProjectRegistryAuthQueries } from "~/projects/data/queries";

import { AppLink, Combobox } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";

export type RegistryOption = { id: string; name: string };

const NONE_REGISTRY_OPTION = {
    value: { id: "", name: "None" },
    label: "none None",
};

/**
 * A pick of the registry an image is pushed to after its build, or none.
 */
export function PushToRegistryCombobox({
    projectId,
    env,
    value,
    onChange,
    readOnly = false,
    invalid = false,
    className,
}: PushToRegistryComboboxProps) {
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

    const comboboxOptions = useMemo(() => {
        const registryOptions = registryAuths.map(auth => {
            const badge = auth.address.trim() || "none";
            return {
                value: { id: auth.id, name: auth.name },
                label: `${badge} ${auth.name}`,
            };
        });

        return [NONE_REGISTRY_OPTION, ...registryOptions];
    }, [registryAuths]);

    return (
        <Combobox
            options={comboboxOptions}
            value={value?.id ?? ""}
            onChange={(_, option) => {
                if (readOnly) {
                    return;
                }

                if (!option || option.id === "") {
                    onChange(undefined);
                    return;
                }

                onChange(option);
            }}
            onSearch={setSearchQuery}
            placeholder="Select registry to push image to"
            searchable
            closeOnSelect
            emptyText="No registry to push image to available"
            className={className}
            valueKey="id"
            aria-invalid={invalid}
            loading={isFetching}
            onRefresh={() => void refetch()}
            isRefreshing={isRefetching}
            splitLabelBadge
            disabled={readOnly}
        />
    );
}

/**
 * Where a project's registry credentials are configured.
 */
export function RegistryCredentialsLink({ projectId }: { projectId: string }) {
    return (
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
    );
}

export interface PushToRegistryComboboxProps {
    projectId: string;
    env: string;
    value: RegistryOption | null | undefined;
    onChange: (value: RegistryOption | undefined) => void;
    readOnly?: boolean;
    invalid?: boolean;
    className?: string;
}
