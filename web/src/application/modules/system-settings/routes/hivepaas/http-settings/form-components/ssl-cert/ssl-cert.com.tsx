import React, { useMemo, useState } from "react";

import { Field, FieldError, FieldGroup } from "@components/ui";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { SslCertQueries } from "~/settings/data";
import { useQuickInstallSslCertDialog } from "~/system-settings/dialogs/quick-install-ssl-cert";

import { AppLink, Combobox, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { HTTP_SETTINGS_TEXT_CONTROL_MAX_WIDTH_CLASS } from "../../http-settings-layout.constants";
import type { HivePaaSHttpSettingsFormInput, HivePaaSHttpSettingsFormOutput } from "../../schemas";

import { SslInfo } from "./ssl-info.com";

function View({ domainIndex, readOnly = false }: SslCertProps) {
    const [searchQuery, setSearchQuery] = useState("");
    const [modalOpen, setModalOpen] = useState(false);
    const [selectedSslId, setSelectedSslId] = useState<string | null>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const { control, setValue } = useFormContext<
        HivePaaSHttpSettingsFormInput,
        unknown,
        HivePaaSHttpSettingsFormOutput
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
    } = SslCertQueries.useFindManyPaginated({
        search: searchQuery,
        domain: normalizedDomain || undefined,
    });

    const { data: sslCertDetail, isFetching: isSslInfoLoading } = SslCertQueries.useFindOneById(
        { id: selectedSslId ?? "" },
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
                        content="TLS certificate for this hostname (global settings scope)."
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
                                        setSelectedSslId(sslCert.value?.id ?? null);
                                        setModalOpen(true);
                                    }}
                                >
                                    Info
                                </button>
                            ) : (
                                <PermissionTooltipAction
                                    id={MODULE_IDS.System}
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

                                                quickInstallActions.open(normalizedDomain);
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
                                to={ROUTE.settings.sslCertificates.$route}
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
