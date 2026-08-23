import { useMemo } from "react";

import { Button } from "@components/ui";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { ChevronDown, Plus } from "lucide-react";
import { useFormContext, useWatch } from "react-hook-form";

import {
    type AppConfigHttpSettingsFormSchemaInput,
    type AppConfigHttpSettingsFormSchemaOutput,
    createDefaultBasicAuthRef,
    createDefaultCircuitBreakerConfig,
    createDefaultClientConfig,
    createDefaultCompressionConfig,
    createDefaultHeaderConfig,
    createDefaultPathRewriteConfig,
    createDefaultRateLimitConfig,
} from "../schemas";

type ConfigKeyDomain =
    | "basicAuth"
    | "clientConfig"
    | "compressionConfig"
    | "headerConfig"
    | "pathRewriteConfig"
    | "rateLimitConfig"
    | "circuitBreakerConfig";
export type ConfigSectionKey = ConfigKeyDomain;

const OPTIONS: { key: ConfigKeyDomain; label: string }[] = [
    { key: "basicAuth", label: "Basic Auth" },
    { key: "clientConfig", label: "Client Configuration" },
    { key: "compressionConfig", label: "Compression Configuration" },
    { key: "headerConfig", label: "Header Configuration" },
    { key: "pathRewriteConfig", label: "Path Rewrite Configuration" },
    { key: "rateLimitConfig", label: "Rate Limit Configuration" },
    { key: "circuitBreakerConfig", label: "Circuit Breaker Configuration" },
];

function segmentAtPath(
    root: AppConfigHttpSettingsFormSchemaInput | undefined,
    basePath: string,
): Record<string, unknown> | undefined {
    if (!root) return undefined;
    const parts = basePath.split(".").filter(Boolean);
    let cur: unknown = root;
    for (const part of parts) {
        if (cur == null || typeof cur !== "object") return undefined;
        cur = (cur as Record<string, unknown>)[part];
    }
    return cur as Record<string, unknown> | undefined;
}

interface AddConfigurationDropdownProps {
    basePath: string;
    readOnly?: boolean;
    onSectionAdded?: (key: ConfigSectionKey) => void;
}

export function AddConfigurationDropdown({
    basePath,
    readOnly = false,
    onSectionAdded,
}: AddConfigurationDropdownProps) {
    const { control, setValue } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const setFormValue = setValue as (name: string, value: unknown, opts?: object) => void;

    const values = useWatch({ control }) as AppConfigHttpSettingsFormSchemaInput | undefined;
    const segment = useMemo(() => segmentAtPath(values, basePath), [values, basePath]);

    const availableOptions = useMemo(() => OPTIONS.filter(({ key }) => segment?.[key] == null), [segment]);

    function add(key: ConfigSectionKey) {
        if (readOnly) {
            return;
        }

        const fieldPath = `${basePath}.${key}`;
        switch (key) {
            case "basicAuth":
                setFormValue(fieldPath, createDefaultBasicAuthRef(), { shouldDirty: true, shouldValidate: true });
                break;
            case "clientConfig":
                setFormValue(fieldPath, createDefaultClientConfig(), { shouldDirty: true, shouldValidate: true });
                break;
            case "compressionConfig":
                setFormValue(fieldPath, createDefaultCompressionConfig(), { shouldDirty: true, shouldValidate: true });
                break;
            case "headerConfig":
                setFormValue(fieldPath, createDefaultHeaderConfig(), { shouldDirty: true, shouldValidate: true });
                break;
            case "pathRewriteConfig":
                setFormValue(fieldPath, createDefaultPathRewriteConfig(), { shouldDirty: true, shouldValidate: true });
                break;
            case "rateLimitConfig":
                setFormValue(fieldPath, createDefaultRateLimitConfig(), { shouldDirty: true, shouldValidate: true });
                break;
            case "circuitBreakerConfig":
                setFormValue(fieldPath, createDefaultCircuitBreakerConfig(), {
                    shouldDirty: true,
                    shouldValidate: true,
                });
                break;
            default:
                break;
        }
        onSectionAdded?.(key);
    }

    if (availableOptions.length === 0) {
        return null;
    }

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button
                    type="button"
                    variant="outline"
                    className="w-fit gap-2 rounded-lg"
                    disabled={readOnly}
                >
                    <Plus className="size-4" />
                    Add Configuration
                    <ChevronDown className="size-4 opacity-60" />
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start">
                {availableOptions.map(({ key, label }) => (
                    <DropdownMenuItem
                        key={key}
                        disabled={readOnly}
                        onSelect={() => {
                            queueMicrotask(() => {
                                add(key);
                            });
                        }}
                    >
                        {label}
                    </DropdownMenuItem>
                ))}
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
