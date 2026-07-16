import { useState } from "react";

import { useFormContext, useWatch } from "react-hook-form";

import { type AppConfigHttpSettingsFormSchemaInput, type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";

import { AddConfigurationDropdown, type ConfigSectionKey } from "./add-configuration-dropdown.com";
import { BasicAuthSection } from "./basic-auth-section.com";
import { CircuitBreakerConfigSection } from "./circuit-breaker-config-section.com";
import { ClientConfigSection } from "./client-config-section.com";
import { CompressionConfigSection } from "./compression-config-section.com";
import { HeaderConfigSection } from "./header-config-section.com";
import { PathRewriteConfigSection } from "./path-rewrite-config-section.com";
import { RateLimitConfigSection } from "./rate-limit-config-section.com";

interface HttpConfigurableSectionsProps {
    basePath: string;
    readOnly?: boolean;
}

export function HttpConfigurableSections({ basePath, readOnly = false }: HttpConfigurableSectionsProps) {
    const { control, setValue, unregister } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();
    const segment = useWatch({ control, name: basePath as never }) as Record<string, unknown> | undefined;
    const setFormValue = setValue as (name: string, value: unknown, opts?: object) => void;
    const [expandSignal, setExpandSignal] = useState<{ key: ConfigSectionKey; seq: number } | null>(null);

    const removeSection = (fieldPath: string) => {
        if (readOnly) {
            return;
        }

        setFormValue(fieldPath, undefined, { shouldDirty: true, shouldValidate: true });
        unregister(fieldPath as never);
    };

    return (
        <div className="flex flex-col gap-4">
            {segment?.["basicAuth"] != null && (
                <BasicAuthSection
                    prefix={`${basePath}.basicAuth`}
                    readOnly={readOnly}
                    onRemove={() => {
                        removeSection(`${basePath}.basicAuth`);
                    }}
                />
            )}

            {segment?.["clientConfig"] != null && (
                <ClientConfigSection
                    prefix={`${basePath}.clientConfig`}
                    readOnly={readOnly}
                    autoExpandToken={expandSignal?.key === "clientConfig" ? expandSignal.seq : undefined}
                    onRemove={() => {
                        removeSection(`${basePath}.clientConfig`);
                    }}
                />
            )}

            {segment?.["compressionConfig"] != null && (
                <CompressionConfigSection
                    prefix={`${basePath}.compressionConfig`}
                    readOnly={readOnly}
                    autoExpandToken={expandSignal?.key === "compressionConfig" ? expandSignal.seq : undefined}
                    onRemove={() => {
                        removeSection(`${basePath}.compressionConfig`);
                    }}
                />
            )}

            {segment?.["headerConfig"] != null && (
                <HeaderConfigSection
                    prefix={`${basePath}.headerConfig`}
                    readOnly={readOnly}
                    autoExpandToken={expandSignal?.key === "headerConfig" ? expandSignal.seq : undefined}
                    onRemove={() => {
                        removeSection(`${basePath}.headerConfig`);
                    }}
                />
            )}

            {segment?.["pathRewriteConfig"] != null && (
                <PathRewriteConfigSection
                    prefix={`${basePath}.pathRewriteConfig`}
                    readOnly={readOnly}
                    autoExpandToken={expandSignal?.key === "pathRewriteConfig" ? expandSignal.seq : undefined}
                    onRemove={() => {
                        removeSection(`${basePath}.pathRewriteConfig`);
                    }}
                />
            )}

            {segment?.["rateLimitConfig"] != null && (
                <RateLimitConfigSection
                    prefix={`${basePath}.rateLimitConfig`}
                    readOnly={readOnly}
                    autoExpandToken={expandSignal?.key === "rateLimitConfig" ? expandSignal.seq : undefined}
                    onRemove={() => {
                        removeSection(`${basePath}.rateLimitConfig`);
                    }}
                />
            )}

            {segment?.["circuitBreakerConfig"] != null && (
                <CircuitBreakerConfigSection
                    prefix={`${basePath}.circuitBreakerConfig`}
                    readOnly={readOnly}
                    autoExpandToken={expandSignal?.key === "circuitBreakerConfig" ? expandSignal.seq : undefined}
                    onRemove={() => {
                        removeSection(`${basePath}.circuitBreakerConfig`);
                    }}
                />
            )}

            <AddConfigurationDropdown
                basePath={basePath}
                readOnly={readOnly}
                onSectionAdded={key => {
                    setExpandSignal(prev => ({ key, seq: (prev?.seq ?? 0) + 1 }));
                }}
            />
        </div>
    );
}
