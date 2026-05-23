import React, { type PropsWithChildren, useCallback, useEffect, useImperativeHandle, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldPath, FormProvider, useForm } from "react-hook-form";
import { type AppContainerSettings } from "~/projects/domain";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";

import {
    GeneralFields,
    HealthcheckFields,
    LabelsFields,
    RestartPolicyFields,
    SecurityFields,
} from "../building-blocks";
import {
    AppConfigContainerSettingsFormSchema,
    type AppConfigContainerSettingsFormSchemaInput,
    type AppConfigContainerSettingsFormSchemaOutput,
    emptyAppConfigContainerSettingsFormDefaults,
} from "../schemas";
import { type AppConfigContainerSettingsFormRef } from "../types";

import { mapAppContainerSettingsToFormInput } from "./app-config-container-settings.form-mappers";

type SchemaInput = AppConfigContainerSettingsFormSchemaInput;
type SchemaOutput = AppConfigContainerSettingsFormSchemaOutput;
type SectionValue = "general" | "labels" | "restart-policy" | "security" | "healthcheck";

const SECTION_BY_FIELD: Record<keyof SchemaInput, SectionValue> = {
    general: "general",
    serviceLabels: "labels",
    containerLabels: "labels",
    restartPolicy: "restart-policy",
    privileges: "security",
    healthcheck: "healthcheck",
};

const CONTAINER_SETTINGS_SECTIONS: {
    value: SectionValue;
    title: string;
    content: React.ReactNode;
}[] = [
    { value: "general", title: "General", content: <GeneralFields /> },
    { value: "labels", title: "Labels", content: <LabelsFields /> },
    { value: "restart-policy", title: "Restart Policy", content: <RestartPolicyFields /> },
    { value: "security", title: "Security", content: <SecurityFields /> },
    { value: "healthcheck", title: "Healthcheck", content: <HealthcheckFields /> },
];

function getSectionValueFromPath(path: string): SectionValue | undefined {
    const [field] = path.split(".");

    return SECTION_BY_FIELD[field as keyof SchemaInput];
}

export function AppConfigContainerSettingsForm({ ref, defaultValues, onSubmit, children }: Props) {
    const methods = useForm<SchemaInput, unknown, SchemaOutput>({
        defaultValues: defaultValues
            ? mapAppContainerSettingsToFormInput(defaultValues)
            : emptyAppConfigContainerSettingsFormDefaults,
        resolver: zodResolver(AppConfigContainerSettingsFormSchema),
        mode: "onSubmit",
    });
    const [openSections, setOpenSections] = useState<SectionValue[]>(["general"]);

    const openSectionsForPaths = useCallback((paths: string[]) => {
        const nextSections = paths.flatMap(path => {
            const section = getSectionValueFromPath(path);

            return section ? [section] : [];
        });

        if (nextSections.length === 0) {
            return;
        }

        setOpenSections(current => Array.from(new Set([...current, ...nextSections])));
    }, []);

    useEffect(() => {
        if (methods.formState.submitCount === 0) {
            return;
        }

        openSectionsForPaths(Object.keys(methods.formState.errors));
    }, [methods.formState.errors, methods.formState.submitCount, openSectionsForPaths]);

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
                        {
                            shouldFocus: index === 0,
                        },
                    );
                });
                openSectionsForPaths(error.errors.map(({ path }) => path));
            },
        }),
        [methods, openSectionsForPaths],
    );

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
                    <Accordion
                        type="multiple"
                        value={openSections}
                        onValueChange={value => {
                            setOpenSections(value as SectionValue[]);
                        }}
                        className="w-full flex flex-col gap-4"
                    >
                        {CONTAINER_SETTINGS_SECTIONS.map(section => (
                            <AccordionItem
                                key={section.value}
                                value={section.value}
                                className="border-none"
                            >
                                <AccordionTrigger className="px-3 py-2 [&>svg]:rotate-90 [&[data-state=open]>svg]:rotate-0 bg-accent">
                                    {section.title}
                                </AccordionTrigger>
                                <AccordionContent className="pt-4 pb-0 pl-4">{section.content}</AccordionContent>
                            </AccordionItem>
                        ))}
                    </Accordion>
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<AppConfigContainerSettingsFormRef>;
    defaultValues?: AppContainerSettings;
    onSubmit: (values: SchemaOutput) => void;
}>;
