import React, { useMemo, useState } from "react";

import { useParams } from "react-router";
import invariant from "tiny-invariant";
import {
    AppConfigFilesQueries,
    ProjectAppSecretsQueries,
    ProjectBasicAuthQueries,
    ProjectSSHKeyQueries,
    ProjectSslCertQueries,
} from "~/projects/data/queries";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { Combobox } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";

/**
 * The settings of one type an entry may mount from. A secret or config file is
 * listed through the app, which shows the project's inheritable ones too; the
 * other types through the environment.
 */
function View({ sourceType, value, onChange, invalid, disabled }: Props) {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");
    const [search, setSearch] = useState("");
    const envRequest = { projectID: projectId, env, search };
    const appRequest = { projectID: projectId, env, appID: appId, search };

    const secrets = ProjectAppSecretsQueries.useFindManyPaginated(appRequest, { enabled: sourceType === "secret" });
    const configFiles = AppConfigFilesQueries.useFindManyPaginated(appRequest, {
        enabled: sourceType === "config-file",
    });
    const certs = ProjectSslCertQueries.useFindManyPaginated(envRequest, { enabled: sourceType === "ssl-cert" });
    const keys = ProjectSSHKeyQueries.useFindManyPaginated(envRequest, { enabled: sourceType === "ssh-key" });
    const auths = ProjectBasicAuthQueries.useFindManyPaginated(envRequest, { enabled: sourceType === "basic-auth" });
    const active = {
        "secret": secrets,
        "config-file": configFiles,
        "ssl-cert": certs,
        "ssh-key": keys,
        "basic-auth": auths,
    }[sourceType];
    const settings: { id: string; name: string }[] = active?.data?.data ?? DEFAULT_PAGINATED_DATA.data;

    const options = useMemo(() => {
        const list = settings.map(setting => ({
            value: { id: setting.id, name: setting.name },
            label: setting.name,
        }));
        if (value?.id && !list.some(item => item.value.id === value.id)) {
            list.unshift({ value: { id: value.id, name: value.name || value.id }, label: value.name || value.id });
        }
        return list;
    }, [settings, value]);

    return (
        <Combobox
            options={options}
            value={value?.id ?? null}
            onChange={(_, option) => {
                onChange(option?.id ? { id: option.id, name: option.name } : null);
            }}
            onSearch={setSearch}
            placeholder="Select a setting"
            searchable
            closeOnSelect
            emptyText="No setting of this type is available to this app"
            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
            valueKey="id"
            aria-invalid={invalid}
            loading={active?.isFetching ?? false}
            onRefresh={() => void active?.refetch()}
            isRefreshing={active?.isRefetching ?? false}
            disabled={disabled}
        />
    );
}

interface Props {
    sourceType: string;
    value: { id: string; name: string } | null;
    onChange: (value: { id: string; name: string } | null) => void;
    invalid?: boolean;
    disabled?: boolean;
}

export const SourcePicker = React.memo(View);
