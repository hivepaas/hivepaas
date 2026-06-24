import React, { useMemo, useState } from "react";

import { Field, FieldError, FieldGroup } from "@components/ui";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { Link, useParams } from "react-router";
import invariant from "tiny-invariant";
import { useQuickInstallSslCertDialog } from "~/projects/dialogs/quick-install-ssl-cert";

import { Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { ProjectSslCertQueries } from "@application/modules/projects/data";

import { HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS } from "../../http-settings-layout.constants";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../../schemas";

import { SslInfo } from "./ssl-info.com";

function View({ domainIndex, readOnly = false }: SslCertProps) {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

    const [searchQuery, setSearchQuery] = useState("");
    const [modalOpen, setModalOpen] = useState(false);
    const [selectedSslId, setSelectedSslId] = useState<string | null>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const { control, setValue } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();
    const domainValue = useWatch({ control, name: `domains.${domainIndex}.domain` });
    const normalizedDomain = typeof domainValue === "string" ? domainValue.trim() : "";

    const {
        field: sslCert,
        fieldState: { error: sslCertError, invalid: isSslCertInvalid },
    } = useController({ control, name: `domains.${domainIndex}.sslCert` });

    const { actions: quickInstallActions } = useQuickInstallSslCertDialog({
        onSuccess: created => {
            setSelectedSslId(created.id);
            setValue(`domains.${domainIndex}.sslCert`, { id: created.id, name: created.name }, { shouldDirty: true });
            void refetch();
        },
    });

    const {
        data: { data: sslCerts } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = ProjectSslCertQueries.useFindManyPaginated({
        projectID: projectId,
        search: searchQuery,
        domain: normalizedDomain || undefined,
    });

    const { data: sslCertDetail, isFetching: isSslInfoLoading } = ProjectSslCertQueries.useFindOneById(
        { projectID: projectId, id: selectedSslId ?? "" },
        {
            enabled: Boolean(selectedSslId),
        },
    );

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
                                    if (readOnly) {
                                        return;
                                    }

                                    if (!option) {
                                        setValue(`domains.${domainIndex}.sslCert`, undefined, { shouldDirty: true });
                                        setSelectedSslId(null);
                                        setModalOpen(false);
                                        return;
                                    }

                                    setValue(
                                        `domains.${domainIndex}.sslCert`,
                                        { id: option.id, name: option.name },
                                        { shouldDirty: true },
                                    );
                                    setSelectedSslId(option.id);
                                }}
                                onSearch={setSearchQuery}
                                placeholder="Select SSL certificate"
                                searchable
                                closeOnSelect
                                emptyText="No SSL certificates available"
                                className={HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS}
                                valueKey="id"
                                aria-invalid={isSslCertInvalid}
                                loading={isFetching}
                                onRefresh={() => void refetch()}
                                isRefreshing={isRefetching}
                                splitLabelBadge
                                allowClear
                                disabled={readOnly}
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
                                <PermissionTooltipAction
                                    id={MODULE_IDS.Project}
                                    action="write"
                                    triggerClassName="inline-flex"
                                >
                                    {({ isDenied }) => (
                                        <button
                                            type="button"
                                            className="text-blue-500 cursor-pointer hover:underline select-none disabled:opacity-50 disabled:cursor-not-allowed"
                                            disabled={isDenied || readOnly}
                                            onClick={() => {
                                                if (!canWrite || readOnly) {
                                                    return;
                                                }

                                                quickInstallActions.open(projectId, normalizedDomain);
                                            }}
                                        >
                                            Quick Install
                                        </button>
                                    )}
                                </PermissionTooltipAction>
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
    readOnly?: boolean;
}

export const SslCert = React.memo(View);
