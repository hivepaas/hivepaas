import { useMemo } from "react";

import { Button } from "@components/ui";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from "@components/ui/dropdown-menu";
import { ChevronDown, Plus } from "lucide-react";
import { useFormContext, useWatch } from "react-hook-form";

import {
    createDefaultBasicAuthRef,
    createDefaultClientConfig,
    createDefaultCompressionConfig,
    createDefaultHeaderConfig,
    createDefaultRateLimitConfig,
    type AppConfigHttpSettingsFormSchemaInput,
    type AppConfigHttpSettingsFormSchemaOutput,
} from "../schemas";

export type HttpConfigSectionScope = "domain" | "path";

type ConfigKeyDomain = "basicAuth" | "clientConfig" | "compressionConfig" | "headerConfig" | "rateLimitConfig";
type ConfigKeyPath = "basicAuth" | "clientConfig" | "rateLimitConfig";

const DOMAIN_OPTIONS: { key: ConfigKeyDomain; label: string }[] = [
    { key: "basicAuth", label: "Basic Auth" },
    { key: "clientConfig", label: "Client Configuration" },
    { key: "compressionConfig", label: "Compression Configuration" },
    { key: "headerConfig", label: "Header Configuration" },
    { key: "rateLimitConfig", label: "Rate Limit Configuration" },
];

const PATH_OPTIONS: { key: ConfigKeyPath; label: string }[] = [
    { key: "basicAuth", label: "Basic Auth" },
    { key: "clientConfig", label: "Client Configuration" },
    { key: "rateLimitConfig", label: "Rate Limit Configuration" },
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
    scope: HttpConfigSectionScope;
}

export function AddConfigurationDropdown({ basePath, scope }: AddConfigurationDropdownProps) {
    const { control, setValue } = useFormContext<
        AppConfigHttpSettingsFormSchemaInput,
        unknown,
        AppConfigHttpSettingsFormSchemaOutput
    >();

    const setFormValue = setValue as (name: string, value: unknown, opts?: object) => void;

    const values = useWatch({ control }) as AppConfigHttpSettingsFormSchemaInput | undefined;
    const segment = useMemo(() => segmentAtPath(values, basePath), [values, basePath]);

    const options = scope === "domain" ? DOMAIN_OPTIONS : PATH_OPTIONS;

    function add(key: ConfigKeyDomain | ConfigKeyPath) {
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
            case "rateLimitConfig":
                setFormValue(fieldPath, createDefaultRateLimitConfig(), { shouldDirty: true, shouldValidate: true });
                break;
            default:
                break;
        }
    }

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button
                    type="button"
                    variant="outline"
                    className="w-fit gap-2 rounded-lg"
                >
                    <Plus className="size-4" />
                    Add Configuration
                    <ChevronDown className="size-4 opacity-60" />
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start">
                {options.map(({ key, label }) => {
                    const present = segment?.[key] != null;
                    return (
                        <DropdownMenuItem
                            key={key}
                            disabled={present}
                            onSelect={e => {
                                e.preventDefault();
                                if (!present) add(key);
                            }}
                        >
                            {label}
                        </DropdownMenuItem>
                    );
                })}
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
