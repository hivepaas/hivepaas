import { Button, Field, FieldError, FieldGroup, Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { PasswordInput } from "@/components/ui/input-password";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { type AppConfigKindSettingsFormSchemaInput, type AppConfigKindSettingsFormSchemaOutput } from "../schemas";

import { KindSslCertSelect } from "./kind-ssl-cert-select.com";

interface Props {
    readOnly?: boolean;
}

const MEMORY_PRESETS = ["128mb", "256mb", "512mb", "1gb", "2gb", "4gb", "8gb"];

const EVICTION_RULES = [
    { value: "noeviction", label: "noeviction", description: "Return errors when memory limit is reached" },
    { value: "allkeys-lru", label: "allkeys-lru", description: "Evict least recently used keys out of all keys" },
    { value: "volatile-lru", label: "volatile-lru", description: "Evict least recently used keys with an expire set" },
    { value: "allkeys-lfu", label: "allkeys-lfu", description: "Evict least frequently used keys out of all keys" },
    {
        value: "volatile-lfu",
        label: "volatile-lfu",
        description: "Evict least frequently used keys with an expire set",
    },
    { value: "allkeys-random", label: "allkeys-random", description: "Randomly evict keys out of all keys" },
    { value: "volatile-random", label: "volatile-random", description: "Randomly evict keys with an expire set" },
    { value: "volatile-ttl", label: "volatile-ttl", description: "Evict keys with the shortest time-to-live" },
] as const;

const PERSISTENCE_MODES = [
    { value: "none", label: "None", description: "Pure in-memory only, no persistence" },
    { value: "rdb", label: "RDB Snapshots", description: "Point-in-time snapshots at configured intervals" },
    { value: "aof", label: "Append-Only File (AOF)", description: "Log every write operation received by server" },
    { value: "rdb+aof", label: "RDB + AOF", description: "Combine RDB point-in-time snapshots with AOF logs" },
] as const;

export function CacheKindFields({ readOnly = false }: Props) {
    const { control } = useFormContext<
        AppConfigKindSettingsFormSchemaInput,
        unknown,
        AppConfigKindSettingsFormSchemaOutput
    >();

    const {
        field: password,
        fieldState: { invalid: isPasswordInvalid, error: passwordError },
    } = useController({ control, name: "cache.password" });

    const {
        field: maxMemory,
        fieldState: { invalid: isMaxMemoryInvalid, error: maxMemoryError },
    } = useController({ control, name: "cache.maxMemory" });

    const {
        field: evictionRule,
        fieldState: { error: evictionRuleError },
    } = useController({ control, name: "cache.evictionRule" });

    const {
        field: persistenceMode,
        fieldState: { error: persistenceModeError },
    } = useController({ control, name: "cache.persistenceMode" });

    return (
        <>
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Password / Token"
                        content="Authentication password for the cache service (e.g. Redis requirepass). Leave empty if unconfigured, or keep existing masked value unchanged."
                    />
                }
            >
                <FieldGroup>
                    <Field className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}>
                        <PasswordInput
                            {...password}
                            value={password.value ?? ""}
                            onChange={password.onChange}
                            aria-invalid={isPasswordInvalid}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                            disabled={readOnly}
                        />
                        <FieldError errors={[passwordError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Max Memory"
                        content="Maximum memory capacity limit before cache eviction policy is triggered (e.g. 512mb, 1gb, 2gb)."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <div className="flex flex-col gap-2">
                            <Input
                                {...maxMemory}
                                value={maxMemory.value ?? ""}
                                onChange={maxMemory.onChange}
                                placeholder="e.g. 512mb, 1gb"
                                aria-invalid={isMaxMemoryInvalid}
                                className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                disabled={readOnly}
                            />
                            {!readOnly && (
                                <div className="flex flex-wrap gap-1.5 items-center">
                                    <span className="text-xs text-muted-foreground mr-1">Presets:</span>
                                    {MEMORY_PRESETS.map(preset => (
                                        <Button
                                            key={preset}
                                            type="button"
                                            variant="outline"
                                            size="sm"
                                            onClick={() => {
                                                maxMemory.onChange(preset);
                                            }}
                                            className="h-6 px-2 text-xs"
                                        >
                                            {preset}
                                        </Button>
                                    ))}
                                </div>
                            )}
                        </div>
                        <FieldError errors={[maxMemoryError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Eviction Policy"
                        content="Key eviction algorithm when maximum memory threshold is reached."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Select
                            value={evictionRule.value ?? ""}
                            onValueChange={evictionRule.onChange}
                            disabled={readOnly}
                        >
                            <SelectTrigger className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}>
                                <SelectValue placeholder="Select eviction policy (optional)" />
                            </SelectTrigger>
                            <SelectContent>
                                {EVICTION_RULES.map(rule => (
                                    <SelectItem
                                        key={rule.value}
                                        value={rule.value}
                                    >
                                        <div className="flex flex-col text-left">
                                            <span className="font-mono text-xs font-semibold">{rule.label}</span>
                                            <span className="text-xs text-muted-foreground">{rule.description}</span>
                                        </div>
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                        <FieldError errors={[evictionRuleError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Persistence Mode"
                        content="Storage persistence strategy for in-memory cache."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Select
                            value={persistenceMode.value ?? ""}
                            onValueChange={persistenceMode.onChange}
                            disabled={readOnly}
                        >
                            <SelectTrigger className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}>
                                <SelectValue placeholder="Select persistence mode (optional)" />
                            </SelectTrigger>
                            <SelectContent>
                                {PERSISTENCE_MODES.map(mode => (
                                    <SelectItem
                                        key={mode.value}
                                        value={mode.value}
                                    >
                                        <div className="flex flex-col text-left">
                                            <span className="font-medium text-sm">{mode.label}</span>
                                            <span className="text-xs text-muted-foreground">{mode.description}</span>
                                        </div>
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                        <FieldError errors={[persistenceModeError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <KindSslCertSelect
                name="cache.sslCert"
                readOnly={readOnly}
            />
        </>
    );
}
