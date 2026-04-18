import { useMemo, useState } from "react";

import { Checkbox, Field, FieldError, FieldGroup, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { Link, useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectSslCertQueries } from "~/projects/data/queries";

import { Combobox, InfoBlock, InputNumberWithAddon, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

interface DomainGeneralFieldsProps {
    domainIndex: number;
}

export function DomainGeneralFields({ domainIndex }: DomainGeneralFieldsProps) {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const p = `domains.${domainIndex}` as const;

    const [searchQuery, setSearchQuery] = useState("");

    const {
        data: { data: sslCerts } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = ProjectSslCertQueries.useFindManyPaginated({ projectID: projectId, search: searchQuery });

    const { field: enabled } = useController({ control, name: `${p}.enabled` });
    const {
        field: containerPort,
        fieldState: { error: containerPortError },
    } = useController({ control, name: `${p}.containerPort` });
    const {
        field: domainRedirect,
        fieldState: { error: domainRedirectError },
    } = useController({ control, name: `${p}.domainRedirect` });
    const { field: forceHttps } = useController({ control, name: `${p}.forceHttps` });
    const {
        field: sslCert,
        fieldState: { error: sslCertError, invalid: isSslCertInvalid },
    } = useController({ control, name: `${p}.sslCert` });

    const comboboxOptions = useMemo(() => {
        return sslCerts.map(cert => ({
            value: { id: cert.id, name: cert.name },
            label: cert.name,
        }));
    }, [sslCerts]);

    return (
        <>
            <InfoBlock title="Enabled">
                <Checkbox
                    checked={enabled.value}
                    onCheckedChange={enabled.onChange}
                />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Container Port"
                        isRequired
                    />
                }
            >
                <InputNumberWithAddon
                    addonLeft="Port"
                    value={containerPort.value}
                    onValueChange={v => {
                        containerPort.onChange(v ?? 0);
                    }}
                    useGrouping={false}
                    placeholder="80"
                    classNameContainer="max-w-[240px]"
                />
                <FieldError errors={[containerPortError]} />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="SSL certificate"
                        content="TLS certificate for this hostname (project-scoped)."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Combobox
                            options={comboboxOptions}
                            value={sslCert.value?.id ?? null}
                            onChange={(_, option) => {
                                sslCert.onChange(option ? { id: option.id, name: option.name } : undefined);
                            }}
                            onSearch={setSearchQuery}
                            placeholder="Select SSL certificate"
                            searchable
                            closeOnSelect
                            emptyText="No SSL certificates available"
                            className="max-w-[400px]"
                            valueKey="id"
                            aria-invalid={isSslCertInvalid}
                            loading={isFetching}
                            onRefresh={() => void refetch()}
                            isRefreshing={isRefetching}
                            splitLabelBadge
                            allowClear
                        />
                        <FieldError errors={[sslCertError]} />
                        <div className="text-xs text-muted-foreground">
                            <p>
                                Need another certificate?{" "}
                                <Link
                                    to="#"
                                    className="text-primary underline-offset-4 hover:underline"
                                >
                                    Add in project settings
                                </Link>
                            </p>
                        </div>
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock title="Redirect To">
                <Input
                    {...domainRedirect}
                    value={domainRedirect.value}
                    onChange={domainRedirect.onChange}
                    placeholder="https://other-domain.com"
                    className="max-w-[400px]"
                />
                <FieldError errors={[domainRedirectError]} />
            </InfoBlock>

            <InfoBlock title="Force HTTPS">
                <Checkbox
                    checked={forceHttps.value}
                    onCheckedChange={forceHttps.onChange}
                />
            </InfoBlock>
        </>
    );
}
