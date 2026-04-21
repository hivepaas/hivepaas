import { useFormContext, useWatch } from "react-hook-form";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

import { AddConfigurationDropdown, type HttpConfigSectionScope } from "./add-configuration-dropdown.com";
import { BasicAuthSection } from "./basic-auth-section.com";
import { ClientConfigSection } from "./client-config-section.com";
import { CompressionConfigSection } from "./compression-config-section.com";
import { HeaderConfigSection } from "./header-config-section.com";
import { RateLimitConfigSection } from "./rate-limit-config-section.com";

interface HttpConfigurableSectionsProps {
    basePath: string;
    scope: HttpConfigSectionScope;
}

export function HttpConfigurableSections({ basePath, scope }: HttpConfigurableSectionsProps) {
    const { control, unregister } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();
    const segment = useWatch({ control, name: basePath as never }) as Record<string, unknown> | undefined;

    return (
        <div className="flex flex-col gap-4">
            {segment?.["basicAuth"] != null && (
                <BasicAuthSection
                    prefix={`${basePath}.basicAuth`}
                    onRemove={() => {
                        unregister(`${basePath}.basicAuth` as never);
                    }}
                />
            )}

            {segment?.["clientConfig"] != null && (
                <ClientConfigSection
                    prefix={`${basePath}.clientConfig`}
                    onRemove={() => {
                        unregister(`${basePath}.clientConfig` as never);
                    }}
                />
            )}

            {segment?.["compressionConfig"] != null && (
                <CompressionConfigSection
                    prefix={`${basePath}.compressionConfig`}
                    onRemove={() => {
                        unregister(`${basePath}.compressionConfig` as never);
                    }}
                />
            )}

            {segment?.["headerConfig"] != null && (
                <HeaderConfigSection
                    prefix={`${basePath}.headerConfig`}
                    onRemove={() => {
                        unregister(`${basePath}.headerConfig` as never);
                    }}
                />
            )}

            {segment?.["rateLimitConfig"] != null && (
                <RateLimitConfigSection
                    prefix={`${basePath}.rateLimitConfig`}
                    onRemove={() => {
                        unregister(`${basePath}.rateLimitConfig` as never);
                    }}
                />
            )}

            <AddConfigurationDropdown
                basePath={basePath}
                scope={scope}
            />
        </div>
    );
}
