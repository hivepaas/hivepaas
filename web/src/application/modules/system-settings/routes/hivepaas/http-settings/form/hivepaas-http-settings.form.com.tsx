import React, { type PropsWithChildren, useEffect, useImperativeHandle, useRef, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldPath, FormProvider, useForm, useWatch } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import type { HivePaaSHttpSettings } from "~/system-settings/domain";

import { ContentBlock } from "@application/shared/components";

import type { ValidationException } from "@infrastructure/exceptions/validation";

import { ClientConfigSection, DomainSelector, RateLimitConfigSection } from "../building-blocks";
import { SslCert } from "../form-components/ssl-cert";
import {
    type HivePaaSHttpSettingsFormInput,
    type HivePaaSHttpSettingsFormOutput,
    HivePaaSHttpSettingsFormSchema,
    emptyHivePaaSHttpSettingsFormDefaults,
} from "../schemas";
import type { HivePaaSHttpSettingsFormRef } from "../types";

import { mapHivePaaSHttpSettingsToFormInput } from "./hivepaas-http-settings.form-mappers";

type SchemaInput = HivePaaSHttpSettingsFormInput;
type SchemaOutput = HivePaaSHttpSettingsFormOutput;

function NoteBox({ children }: PropsWithChildren) {
    return <div className={cn(dashedBorderBox, "text-sm leading-6 text-center")}>{children}</div>;
}

function ConditionalDomainDetailSections({
    activeDomainIndex,
    setActiveDomainIndex,
    readOnly,
}: {
    activeDomainIndex: number;
    setActiveDomainIndex: React.Dispatch<React.SetStateAction<number>>;
    readOnly: boolean;
}) {
    const domains = useWatch<SchemaInput, "domains">({ name: "domains" });
    const hasDomains = domains.length > 0;
    const activeDomain = activeDomainIndex >= 0 ? domains[activeDomainIndex] : undefined;
    const hasActiveDomain = Boolean(activeDomain);

    useEffect(() => {
        const len = domains.length;
        if (len === 0) {
            if (activeDomainIndex !== -1) {
                setActiveDomainIndex(-1);
            }
            return;
        }

        if (activeDomainIndex >= len) {
            setActiveDomainIndex(len - 1);
        }
    }, [activeDomainIndex, domains.length, setActiveDomainIndex]);

    if (!hasDomains || !hasActiveDomain) {
        return null;
    }

    return (
        <>
            <ContentBlock label="General">
                <div className="flex flex-col gap-6 px-2">
                    <SslCert
                        domainIndex={activeDomainIndex}
                        readOnly={readOnly}
                    />
                </div>
            </ContentBlock>

            <div className="flex flex-col gap-6 px-2">
                <ClientConfigSection
                    prefix={`domains.${activeDomainIndex}.clientConfig`}
                    readOnly={readOnly}
                />
                <RateLimitConfigSection
                    prefix={`domains.${activeDomainIndex}.rateLimitConfig`}
                    readOnly={readOnly}
                />
            </div>
        </>
    );
}

export function HivePaaSHttpSettingsForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
    const [activeDomainIndex, setActiveDomainIndex] = useState(0);

    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapHivePaaSHttpSettingsToFormInput(defaultValues)
            : emptyHivePaaSHttpSettingsFormDefaults,
        resolver: zodResolver(HivePaaSHttpSettingsFormSchema),
        mode: "onSubmit",
    });

    const activeDomainIndexRef = useRef(activeDomainIndex);
    useEffect(() => {
        activeDomainIndexRef.current = activeDomainIndex;
    }, [activeDomainIndex]);

    useUpdateEffect(() => {
        const prevName = methods.getValues().domains[activeDomainIndexRef.current]?.domain;
        methods.reset(
            defaultValues ? mapHivePaaSHttpSettingsToFormInput(defaultValues) : emptyHivePaaSHttpSettingsFormDefaults,
        );
        const newDomains = defaultValues?.domains ?? [];
        if (newDomains.length === 0) {
            setActiveDomainIndex(-1);
            return;
        }
        const trimmedPrev = prevName?.trim() ?? "";
        const idx = trimmedPrev ? newDomains.findIndex(d => d.domain.trim() === trimmedPrev) : -1;
        setActiveDomainIndex(idx >= 0 ? idx : 0);
    }, [defaultValues]);

    useImperativeHandle(
        ref,
        () => ({
            setValues: (values: Partial<SchemaInput>) => {
                methods.reset({
                    ...methods.getValues(),
                    ...values,
                } as SchemaInput);
            },
            onError(error: ValidationException) {
                if (error.errors.length === 0) {
                    return;
                }

                error.errors.forEach(({ path, message }, index) => {
                    methods.setError(
                        path as FieldPath<SchemaInput>,
                        { message, type: "manual" },
                        { shouldFocus: index === 0 },
                    );
                });
            },
        }),
        [methods],
    );

    return (
        <FormProvider {...methods}>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    if (readOnly) {
                        return;
                    }

                    void methods.handleSubmit(onSubmit)(event);
                }}
                className="flex flex-col gap-6"
            >
                <fieldset
                    disabled={readOnly}
                    className="contents"
                >
                    <NoteBox>
                        <div>
                            <span className="text-orange-500">Note:</span> If you want to change the app&apos;s domain,
                            please add a new domain, set it as default, and verify that it works as expected before
                            deleting the old one.
                        </div>
                        <div className="mt-2">
                            <span className="text-orange-500">Warning:</span> Incorrect configuration may cause you to
                            lose access to the dashboard. These changes may cause the application to restart, and you
                            may lose access to the dashboard for approximately 30 seconds.
                        </div>
                    </NoteBox>

                    <DomainSelector
                        activeDomainIndex={activeDomainIndex}
                        setActiveDomainIndex={setActiveDomainIndex}
                        readOnly={readOnly}
                    />

                    <ConditionalDomainDetailSections
                        activeDomainIndex={activeDomainIndex}
                        setActiveDomainIndex={setActiveDomainIndex}
                        readOnly={readOnly}
                    />
                </fieldset>

                {children}
            </form>
        </FormProvider>
    );
}

interface Props extends PropsWithChildren {
    ref?: React.Ref<HivePaaSHttpSettingsFormRef>;
    defaultValues?: HivePaaSHttpSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}
