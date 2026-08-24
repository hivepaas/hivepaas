import React, { useMemo, useState } from "react";

import { Field, FieldError, FieldGroup } from "@components/ui";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { useQuickInstallSslCertDialog } from "~/projects/dialogs/quick-install-ssl-cert";

import { AppLink, Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { ProjectSslCertQueries } from "@application/modules/projects/data";

import { HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS } from "../../routing-settings-layout.constants";
import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../../schemas";

import { SslInfo } from "./ssl-info.com";

function View({ domainIndex, readOnly = false }: SslCertProps) {
    const { id: projectId, env } = useParams<{ id: string; env: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");

    const [searchQuery, setSearchQuery] = useState("");
    const [modalOpen, setModalOpen] = useState(false);
    const [selectedSslId, setSelectedSslId] = useState<string | null>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const { control } = useFormContext<
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
            sslCert.onChange({ id: created.id, name: created.name });
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
        env,
        search: searchQuery,
        domain: normalizedDomain || undefined,
    });

    const { data: sslCertDetail, isFetching: isSslInfoLoading } = ProjectSslCertQueries.useFindOneById(
        { projectID: projectId, env, id: selectedSslId ?? "" },
        {
            enabled: Boolean(selectedSslId),
        },
    );

    const comboboxOptions = useMemo(() => {
        const list = sslCerts.map(cert => ({
            value: { id: cert.id, name: cert.name },
            label: cert.name,
        }));

        const currentSsl = sslCert.value;
        if (currentSsl?.id && !list.some(item => item.value.id.toLowerCase() === currentSsl.id.toLowerCase())) {
            list.unshift({
                value: { id: currentSsl.id, name: currentSsl.name || currentSsl.id },
                label: currentSsl.name || currentSsl.id,
            });
        }

        return list;
    }, [sslCerts, sslCert.value]);

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

                                    if (!option?.id) {
                                        sslCert.onChange(null);
                                        setSelectedSslId(null);
                                        setModalOpen(false);
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

                                                quickInstallActions.open(projectId, env, normalizedDomain);
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
                            <AppLink.Basic
                                to={ROUTE.projects.single.providerConfiguration.sslCertificates.$route(projectId)}
                                className="text-primary underline-offset-4 hover:underline"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                Configure SSL Certificates
                            </AppLink.Basic>
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
