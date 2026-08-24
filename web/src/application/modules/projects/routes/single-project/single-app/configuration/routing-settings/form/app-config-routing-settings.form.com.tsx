import React, { type PropsWithChildren, useEffect, useImperativeHandle, useRef, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, type FieldPath, FormProvider, useController, useForm, useWatch } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { type AppRoutingSettings } from "~/projects/domain";
import { ERoutingProtocol } from "~/projects/module-shared/enums";

import { ContentBlock } from "@application/shared/components";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import {
    DomainConfigurableSections,
    DomainGeneralFields,
    DomainSelector,
    LBConfigSection,
    PathsSection,
} from "../building-blocks";
import {
    AppConfigRoutingSettingsFormSchema,
    type AppConfigRoutingSettingsFormSchemaInput,
    type AppConfigRoutingSettingsFormSchemaOutput,
    emptyAppConfigRoutingSettingsFormDefaults,
} from "../schemas";
import { type AppConfigRoutingSettingsFormRef } from "../types";

import { mapAppRoutingSettingsToFormInput } from "./app-config-routing-settings.form-mappers";

type SchemaInput = AppConfigRoutingSettingsFormSchemaInput;
type SchemaOutput = AppConfigRoutingSettingsFormSchemaOutput;

function ConditionalDomainDetailSections({
    activeDomainIndex,
    setActiveDomainIndex,
    suppressDomainAutoSelectRef,
    readOnly,
}: {
    activeDomainIndex: number;
    setActiveDomainIndex: React.Dispatch<React.SetStateAction<number>>;
    suppressDomainAutoSelectRef: React.RefObject<boolean>;
    readOnly: boolean;
}) {
    const domains = useWatch<SchemaInput, "domains">({ name: "domains" });
    const hasDomains = domains.length > 0;
    const activeDomain = activeDomainIndex >= 0 ? domains[activeDomainIndex] : undefined;
    const hasActiveDomain = Boolean(activeDomain);
    const activeDomainRedirect = typeof activeDomain?.domainRedirect === "string" ? activeDomain.domainRedirect : "";
    const hasRedirect = Boolean(activeDomainRedirect.trim());
    const isHttp = (activeDomain?.protocol ?? ERoutingProtocol.HTTP) === ERoutingProtocol.HTTP;
    const isTlsPassthrough = Boolean(activeDomain?.tlsPassthrough);
    const showHttpAdvancedSections = isHttp && !isTlsPassthrough;

    useEffect(() => {
        const len = domains.length;
        if (len === 0) {
            // Pending first append: index 0 targets the slot before useWatch reflects it.
            if (activeDomainIndex === 0) {
                return;
            }
            if (activeDomainIndex !== -1) {
                setActiveDomainIndex(-1);
            }
            return;
        }

        // Delete flow briefly sets -1 with suppressDomainAutoSelectRef to unmount detail controllers.
        if (activeDomainIndex === -1) {
            if (!suppressDomainAutoSelectRef.current) {
                setActiveDomainIndex(len - 1);
            }
            return;
        }

        // Pending append: index targets the new slot before useWatch reflects the added domain.
        if (activeDomainIndex === len) {
            return;
        }

        if (activeDomainIndex > len) {
            setActiveDomainIndex(len - 1);
        }
    }, [activeDomainIndex, domains.length, setActiveDomainIndex, suppressDomainAutoSelectRef]);

    if (!hasDomains || !hasActiveDomain) {
        return null;
    }

    return (
        <>
            <div className="flex flex-col gap-6 px-2">
                <DomainGeneralFields
                    domainIndex={activeDomainIndex}
                    readOnly={readOnly}
                />
                {showHttpAdvancedSections && (
                    <LBConfigSection
                        prefix={`domains.${activeDomainIndex}.lbConfig`}
                        readOnly={readOnly}
                    />
                )}
            </div>
            {showHttpAdvancedSections && !hasRedirect && (
                <>
                    <div className="flex flex-col gap-6 px-2">
                        <DomainConfigurableSections
                            domainIndex={activeDomainIndex}
                            readOnly={readOnly}
                        />
                    </div>

                    <ContentBlock label={<span className="text-red-400">Path Configuration</span>}>
                        <div className="flex flex-col gap-6">
                            <PathsSection
                                domainIndex={activeDomainIndex}
                                readOnly={readOnly}
                            />
                        </div>
                    </ContentBlock>
                </>
            )}
        </>
    );
}

export function AppConfigRoutingSettingsForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
    const [activeDomainIndex, setActiveDomainIndex] = useState(0);
    const suppressDomainAutoSelectRef = useRef(false);
    const setSuppressDomainAutoSelect = (value: boolean) => {
        suppressDomainAutoSelectRef.current = value;
    };

    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapAppRoutingSettingsToFormInput(defaultValues)
            : emptyAppConfigRoutingSettingsFormDefaults,
        resolver: zodResolver(AppConfigRoutingSettingsFormSchema),
        mode: "onSubmit",
    });

    const { control } = methods;

    const activeDomainIndexRef = useRef(activeDomainIndex);
    useEffect(() => {
        activeDomainIndexRef.current = activeDomainIndex;
    }, [activeDomainIndex]);

    useUpdateEffect(() => {
        const prevName = methods.getValues().domains[activeDomainIndexRef.current]?.domain;
        methods.reset(
            defaultValues ? mapAppRoutingSettingsToFormInput(defaultValues) : emptyAppConfigRoutingSettingsFormDefaults,
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

    const { field: exposePublicly } = useController({ control, name: "exposePublicly" });

    function onInvalid(errors: FieldErrors<SchemaInput>) {
        console.error(errors);
    }

    return (
        <div className="pt-2">
            <FormProvider {...methods}>
                <form
                    onSubmit={event => {
                        event.preventDefault();
                        if (readOnly) {
                            return;
                        }

                        void methods.handleSubmit(onSubmit, onInvalid)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <fieldset
                        disabled={readOnly}
                        className="contents"
                    >
                        <div className="flex flex-col gap-6 px-2">
                            <DomainSelector
                                activeDomainIndex={activeDomainIndex}
                                setActiveDomainIndex={setActiveDomainIndex}
                                domainSuggestion={defaultValues?.domainSuggestion ?? ""}
                                setSuppressDomainAutoSelect={setSuppressDomainAutoSelect}
                                readOnly={readOnly}
                            />
                        </div>

                        {exposePublicly.value && (
                            <ConditionalDomainDetailSections
                                activeDomainIndex={activeDomainIndex}
                                setActiveDomainIndex={setActiveDomainIndex}
                                suppressDomainAutoSelectRef={suppressDomainAutoSelectRef}
                                readOnly={readOnly}
                            />
                        )}

                        {children}
                    </fieldset>
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigRoutingSettingsFormRef>;
    defaultValues?: AppRoutingSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;

export { AppConfigRoutingSettingsForm as AppConfigHttpSettingsForm };
