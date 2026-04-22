import { useFormContext, useWatch } from "react-hook-form";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

import { AddConfigurationDropdown } from "./add-configuration-dropdown.com";
import { BasicAuthSection } from "./basic-auth-section.com";
import { ClientConfigSection } from "./client-config-section.com";
import { CompressionConfigSection } from "./compression-config-section.com";
import { HeaderConfigSection } from "./header-config-section.com";
import { RateLimitConfigSection } from "./rate-limit-config-section.com";

interface HttpConfigurableSectionsProps {
    basePath: string;
}

export function HttpConfigurableSections({ basePath }: HttpConfigurableSectionsProps) {
    const { control, setValue, unregister } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();
    const segment = useWatch({ control, name: basePath as never }) as Record<string, unknown> | undefined;
    const setFormValue = setValue as (name: string, value: unknown, opts?: object) => void;

    const removeSection = (fieldPath: string) => {
        setFormValue(fieldPath, undefined, { shouldDirty: true, shouldValidate: true });
        unregister(fieldPath as never);
    };

    return (
        <div className="flex flex-col gap-4">
            {segment?.["basicAuth"] != null && (
                <BasicAuthSection
                    prefix={`${basePath}.basicAuth`}
                    onRemove={() => {
                        removeSection(`${basePath}.basicAuth`);
                    }}
                />
            )}

            {segment?.["clientConfig"] != null && (
                <ClientConfigSection
                    prefix={`${basePath}.clientConfig`}
                    onRemove={() => {
                        removeSection(`${basePath}.clientConfig`);
                    }}
                />
            )}

            {segment?.["compressionConfig"] != null && (
                <CompressionConfigSection
                    prefix={`${basePath}.compressionConfig`}
                    onRemove={() => {
                        removeSection(`${basePath}.compressionConfig`);
                    }}
                />
            )}

            {segment?.["headerConfig"] != null && (
                <HeaderConfigSection
                    prefix={`${basePath}.headerConfig`}
                    onRemove={() => {
                        removeSection(`${basePath}.headerConfig`);
                    }}
                />
            )}

            {segment?.["rateLimitConfig"] != null && (
                <RateLimitConfigSection
                    prefix={`${basePath}.rateLimitConfig`}
                    onRemove={() => {
                        removeSection(`${basePath}.rateLimitConfig`);
                    }}
                />
            )}

            <AddConfigurationDropdown basePath={basePath} />
        </div>
    );
}
