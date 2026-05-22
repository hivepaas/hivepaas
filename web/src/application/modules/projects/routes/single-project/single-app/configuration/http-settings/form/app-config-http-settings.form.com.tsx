import React, { type PropsWithChildren, useEffect, useImperativeHandle, useRef, useState } from "react";

import { Button } from "@components/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { X } from "lucide-react";
import { type FieldPath, FormProvider, useController, useFieldArray, useForm, useWatch } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { type AppHttpSettings } from "~/projects/domain";

import { ContentBlock } from "@application/shared/components";
import { PopConfirm } from "@application/shared/components/pop-confirm";

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
    onRemoveDomain,
}: {
    activeDomainIndex: number;
    setActiveDomainIndex: React.Dispatch<React.SetStateAction<number>>;
    onRemoveDomain: (index: number) => void;
}) {
    const domains = useWatch<SchemaInput, "domains">({ name: "domains" });
    const hasDomains = domains.length > 0;
    const activeDomain = domains[activeDomainIndex];
    const hasRedirect = Boolean(activeDomain ? activeDomain.domainRedirect.trim() : "");

    useEffect(() => {
        const len = domains.length;
        if (len === 0) {
            setActiveDomainIndex(-1);
        }
    }, [domains.length, setActiveDomainIndex]);

    if (!hasDomains || activeDomainIndex < 0) {
        return null;
    }

    return (
        <>
            <h3 className="font-medium bg-accent py-2 px-3 rounded-lg text-red-500 flex items-center justify-between">
                Selected Domain: {domains[activeDomainIndex]?.domain ?? ""}
                <PopConfirm
                    title="Remove domain"
                    description="Confirm deletion of this domain?"
                    confirmText="Remove"
                    cancelText="Cancel"
                    variant="destructive"
                    onConfirm={() => {
                        onRemoveDomain(activeDomainIndex);
                    }}
                >
                    <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="shrink-0 text-muted-foreground hover:text-destructive h-0"
                        title="Remove domain"
                    >
                        <X className="size-4" />
                    </Button>
                </PopConfirm>
            </h3>
            <div className="flex flex-col gap-6 px-2">
                <DomainGeneralFields domainIndex={activeDomainIndex} />
                <LBConfigSection prefix={`domains.${activeDomainIndex}.lbConfig`} />
            </div>
            {!hasRedirect && (
                <>
                    <div className="flex flex-col gap-6 px-2">
                        <DomainConfigurableSections domainIndex={activeDomainIndex} />
                    </div>

                    <ContentBlock label="Path Configuration">
                        <div className="flex flex-col gap-6">
                            <PathsSection domainIndex={activeDomainIndex} />
                        </div>
                    </ContentBlock>
                </>
            )}
        </>
    );
}

export function AppConfigHttpSettingsForm({ ref, defaultValues, onSubmit, children }: Props) {
    const [activeDomainIndex, setActiveDomainIndex] = useState(0);

    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapAppHttpSettingsToFormInput(defaultValues)
            : emptyAppConfigHttpSettingsFormDefaults,
        resolver: zodResolver(AppConfigHttpSettingsFormSchema),
        mode: "onSubmit",
    });

    const { control } = methods;
    const { remove } = useFieldArray({ control, name: "domains" });

    const handleRemoveDomain = (index: number) => {
        const before = methods.getValues().domains.length;
        remove(index);
        const after = before - 1;
        setActiveDomainIndex(after > 0 ? 0 : -1);
    };

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

    return (
        <div className="pt-2">
            <FormProvider {...methods}>
                <form
                    onSubmit={event => {
                        event.preventDefault();
                        void methods.handleSubmit(onSubmit)(event);
                    }}
                    className="flex flex-col gap-6"
                >
                    <div className="flex flex-col gap-6 px-2">
                        <DomainSelector
                            activeDomainIndex={activeDomainIndex}
                            setActiveDomainIndex={setActiveDomainIndex}
                            internalEndpoints={defaultValues?.internalEndpoints ?? []}
                            domainSuggestion={defaultValues?.domainSuggestion ?? ""}
                        />
                    </div>

                    {exposePublicly.value && (
                        <ConditionalDomainDetailSections
                            activeDomainIndex={activeDomainIndex}
                            setActiveDomainIndex={setActiveDomainIndex}
                            onRemoveDomain={handleRemoveDomain}
                        />
                    )}

                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigHttpSettingsFormRef>;
    defaultValues?: AppHttpSettings;
    onSubmit: (values: SchemaOutput) => void;
}>;
