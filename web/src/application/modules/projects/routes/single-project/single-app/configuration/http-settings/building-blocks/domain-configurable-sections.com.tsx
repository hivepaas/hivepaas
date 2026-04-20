import { useFormContext, useWatch } from "react-hook-form";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

import { AddConfigurationDropdown } from "./add-configuration-dropdown.com";
import { BasicAuthSection } from "./basic-auth-section.com";
import { ClientConfigSection } from "./client-config-section.com";
import { CompressionConfigSection } from "./compression-config-section.com";
import { HeaderConfigSection } from "./header-config-section.com";
import { RateLimitConfigSection } from "./rate-limit-config-section.com";

interface DomainConfigurableSectionsProps {
    domainIndex: number;
}

export function DomainConfigurableSections({ domainIndex }: DomainConfigurableSectionsProps) {
    const { control, unregister } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const basePath = `domains.${domainIndex}`;

    const basicAuth = useWatch({ control, name: `${basePath}.basicAuth` as never });
    const clientConfig = useWatch({ control, name: `${basePath}.clientConfig` as never });
    const compressionConfig = useWatch({ control, name: `${basePath}.compressionConfig` as never });
    const headerConfig = useWatch({ control, name: `${basePath}.headerConfig` as never });
    const rateLimitConfig = useWatch({ control, name: `${basePath}.rateLimitConfig` as never });

    return (
        <div className="flex flex-col gap-4">
            <AddConfigurationDropdown
                basePath={basePath}
                scope="domain"
            />
            <BasicAuthSection
                prefix={`${basePath}.basicAuth`}
                onRemove={() => {
                    unregister(`${basePath}.basicAuth` as never);
                }}
            />

            <ClientConfigSection
                prefix={`${basePath}.clientConfig`}
                onRemove={() => {
                    unregister(`${basePath}.clientConfig` as never);
                }}
            />

            <CompressionConfigSection
                prefix={`${basePath}.compressionConfig`}
                onRemove={() => {
                    unregister(`${basePath}.compressionConfig` as never);
                }}
            />

            <HeaderConfigSection
                prefix={`${basePath}.headerConfig`}
                onRemove={() => {
                    unregister(`${basePath}.headerConfig` as never);
                }}
            />

            <RateLimitConfigSection
                prefix={`${basePath}.rateLimitConfig`}
                onRemove={() => {
                    unregister(`${basePath}.rateLimitConfig` as never);
                }}
            />
        </div>
    );
}
