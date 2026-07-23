import React, { type PropsWithChildren, useEffect, useImperativeHandle, useRef, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, type FieldPath, FormProvider, useController, useForm, useWatch } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { type AppHttpSettings } from "~/projects/domain";

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
    AppConfigHttpSettingsFormSchema,
    type AppConfigHttpSettingsFormSchemaInput,
    type AppConfigHttpSettingsFormSchemaOutput,
    emptyAppConfigHttpSettingsFormDefaults,
} from "../schemas";
import { type AppConfigHttpSettingsFormRef } from "../types";

import { mapAppHttpSettingsToFormInput } from "./app-config-http-settings.form-mappers";

type SchemaInput = AppConfigHttpSettingsFormSchemaInput;
type SchemaOutput = AppConfigHttpSettingsFormSchemaOutput;

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
    const activeDomainRedirect = typeof activeDomain?.domainRedirect === "string" ? activeDomain.domainRedirect : "";
    const hasRedirect = Boolean(activeDomainRedirect.trim());

    useEffect(() => {
        const len = domains.length;
        if (len === 0) {
            if (activeDomainIndex !== -1) {
                setActiveDomainIndex(-1);
            }
            return;
        }

        // Do not auto-select when index is -1: delete flow unmounts detail controllers
        // before splice, then restores the active index explicitly.
        if (activeDomainIndex >= len) {
            setActiveDomainIndex(len - 1);
        }
    }, [activeDomainIndex, domains.length, setActiveDomainIndex]);

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
                <LBConfigSection
                    prefix={`domains.${activeDomainIndex}.lbConfig`}
                    readOnly={readOnly}
                />
            </div>
            {!hasRedirect && (
                <>
                    <div className="flex flex-col gap-6 px-2">
                        <DomainConfigurableSections
                            domainIndex={activeDomainIndex}
                            readOnly={readOnly}
                        />
                    </div>

                    <ContentBlock label="Path Configuration">
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

export function AppConfigHttpSettingsForm({ ref, defaultValues, onSubmit, readOnly = false, children }: Props) {
    const [activeDomainIndex, setActiveDomainIndex] = useState(0);

    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapAppHttpSettingsToFormInput(defaultValues)
            : emptyAppConfigHttpSettingsFormDefaults,
        resolver: zodResolver(AppConfigHttpSettingsFormSchema),
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
            defaultValues ? mapAppHttpSettingsToFormInput(defaultValues) : emptyAppConfigHttpSettingsFormDefaults,
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
                                readOnly={readOnly}
                            />
                        </div>

                        {exposePublicly.value && (
                            <ConditionalDomainDetailSections
                                activeDomainIndex={activeDomainIndex}
                                setActiveDomainIndex={setActiveDomainIndex}
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
    ref?: React.Ref<AppConfigHttpSettingsFormRef>;
    defaultValues?: AppHttpSettings;
    onSubmit: (values: SchemaOutput) => void;
    readOnly?: boolean;
}>;
