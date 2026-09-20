import { useMemo, useState } from "react";

import { Field, FieldError, FieldGroup } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectSslCertQueries } from "~/projects/data";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";

import { type AppConfigKindSettingsFormSchemaInput, type AppConfigKindSettingsFormSchemaOutput } from "../schemas";

interface Props {
    name: "database.sslCert" | "cache.sslCert";
    readOnly?: boolean;
}

export function KindSslCertSelect({ name, readOnly = false }: Props) {
    const { id: projectId, env } = useParams<{ id: string; env: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");

    const [searchQuery, setSearchQuery] = useState("");

    const { control } = useFormContext<
        AppConfigKindSettingsFormSchemaInput,
        unknown,
        AppConfigKindSettingsFormSchemaOutput
    >();

    const {
        field: sslCert,
        fieldState: { error },
    } = useController({ control, name });

    const { data: { data: sslCerts } = DEFAULT_PAGINATED_DATA, isFetching } =
        ProjectSslCertQueries.useFindManyPaginated({
            projectID: projectId,
            env,
            search: searchQuery,
        });

    const comboboxOptions = useMemo(() => {
        const list = sslCerts.map(cert => ({
            value: { id: cert.id, name: cert.name },
            label: cert.name,
        }));

        const currentSsl = sslCert.value;
        if (currentSsl?.id && !list.some(item => item.value.id.toLowerCase() === currentSsl.id.toLowerCase())) {
            list.unshift({
                value: { id: currentSsl.id, name: currentSsl.name ?? currentSsl.id },
                label: currentSsl.name ?? currentSsl.id,
            });
        }

        return list;
    }, [sslCerts, sslCert.value]);

    return (
        <InfoBlock
            titleWidth={220}
            title={
                <LabelWithInfo
                    label="SSL Certificate"
                    content="Optional TLS certificate from project settings to associate with this service."
                />
            }
        >
            <FieldGroup>
                <Field>
                    <div className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}>
                        <Combobox
                            options={comboboxOptions}
                            value={sslCert.value?.id ?? null}
                            onChange={(_, option) => {
                                if (readOnly) return;
                                if (!option?.id) {
                                    sslCert.onChange(null);
                                    return;
                                }
                                sslCert.onChange({ id: option.id, name: option.name });
                            }}
                            onSearch={setSearchQuery}
                            placeholder="Select SSL certificate (optional)"
                            loading={isFetching}
                            disabled={readOnly}
                        />
                    </div>
                    <FieldError errors={[error]} />
                </Field>
            </FieldGroup>
        </InfoBlock>
    );
}
