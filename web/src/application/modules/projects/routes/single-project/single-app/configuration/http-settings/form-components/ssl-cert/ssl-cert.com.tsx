import React, { useMemo, useState } from "react";

import { Field, FieldError, FieldGroup } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { Link, useParams } from "react-router";
import invariant from "tiny-invariant";

import { Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";

import { ProjectSslCertQueries } from "@application/modules/projects/data";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../../schemas";

import { SslInfo } from "./ssl-info.com";

function View({ domainIndex }: SslCertProps) {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

    const [searchQuery, setSearchQuery] = useState("");
    const [modalOpen, setModalOpen] = useState(false);
    const [selectedSslId, setSelectedSslId] = useState<string | null>(null);

    const { control } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const {
        field: sslCert,
        fieldState: { error: sslCertError, invalid: isSslCertInvalid },
    } = useController({ control, name: `domains.${domainIndex}.sslCert` });

    const {
        data: { data: sslCerts } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = ProjectSslCertQueries.useFindManyPaginated({ projectID: projectId, search: searchQuery });

    const { data: sslCertDetail, isFetching: isSslInfoLoading } = ProjectSslCertQueries.useFindOneById(
        { projectID: projectId, id: selectedSslId ?? "" },
        {
            enabled: Boolean(selectedSslId),
        },
    );

    console.log("sslCert", sslCert.value);

    const comboboxOptions = useMemo(() => {
        return sslCerts.map(cert => ({
            value: { id: cert.id, name: cert.name },
            label: cert.name,
        }));
    }, [sslCerts]);

    return (
        <>
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="SSL Certificate"
                        content="TLS certificate for this hostname (project-scoped)."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <div className="flex items-center gap-2">
                            <Combobox
                                options={comboboxOptions}
                                value={sslCert.value?.id ?? null}
                                onChange={(_, option) => {
                                    console.log("option", option);
                                    if (!option) {
                                        sslCert.onChange(null);
                                        setSelectedSslId(null);
                                        return;
                                    }

                                    sslCert.onChange({ id: option.id, name: option.name });
                                    setSelectedSslId(option.id);
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

                            {sslCert.value?.id ? (
                                <button
                                    type="button"
                                    className="text-blue-500 cursor-pointer hover:underline select-none disabled:opacity-50 disabled:cursor-not-allowed"
                                    onClick={() => {
                                        setModalOpen(true);
                                    }}
                                >
                                    Info
                                </button>
                            ) : (
                                <button
                                    type="button"
                                    className="text-blue-500 cursor-pointer hover:underline select-none disabled:opacity-50 disabled:cursor-not-allowed"
                                    onClick={() => {
                                        setModalOpen(true);
                                    }}
                                >
                                    Quick Install
                                </button>
                            )}
                        </div>
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

            <SslInfo
                open={modalOpen}
                onOpenChange={setModalOpen}
                sslCert={sslCertDetail?.data}
                isLoading={isSslInfoLoading}
            />
        </>
    );
}

interface SslCertProps {
    domainIndex: number;
}

export const SslCert = React.memo(View);
