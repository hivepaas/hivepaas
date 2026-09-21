import { useMemo } from "react";

import { Field, FieldError, FieldGroup, Input } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { EditableCombobox, InfoBlock, LabelWithInfo } from "@application/shared/components";

import { APP_KIND_ENGINE_SUGGESTIONS } from "../constants";
import { type AppConfigKindSettingsFormSchemaInput, type AppConfigKindSettingsFormSchemaOutput } from "../schemas";

interface Props {
    readOnly?: boolean;
}

export function EngineFields({ readOnly = false }: Props) {
    const { control } = useFormContext<
        AppConfigKindSettingsFormSchemaInput,
        unknown,
        AppConfigKindSettingsFormSchemaOutput
    >();

    const category = useWatch({ control, name: "category" });

    const {
        field: engine,
        fieldState: { invalid: isEngineInvalid, error: engineError },
    } = useController({ control, name: "engine" });

    const {
        field: version,
        fieldState: { invalid: isVersionInvalid, error: versionError },
    } = useController({ control, name: "version" });

    const {
        field: port,
        fieldState: { invalid: isPortInvalid, error: portError },
    } = useController({ control, name: "port" });

    const engineSuggestions = useMemo(() => {
        return (APP_KIND_ENGINE_SUGGESTIONS as Record<string, string[]>)[category] ?? [];
    }, [category]);

    const enginePlaceholder = useMemo(() => {
        switch (category) {
            case "database":
                return "e.g. postgres, mysql, mongodb";
            case "cache":
                return "e.g. redis, memcached, valkey";
            case "storage":
                return "e.g. minio, seaweedfs, s3-compatible";
            default:
                return "e.g. nodejs, python, nextjs, go";
        }
    }, [category]);

    return (
        <>
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Engine"
                        isRequired
                        content="The software engine or runtime running in the container (e.g. postgres, mysql, redis, node, python)."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <EditableCombobox
                            options={engineSuggestions}
                            value={engine.value}
                            onChange={engine.onChange}
                            placeholder={enginePlaceholder}
                            aria-invalid={isEngineInvalid}
                            disabled={readOnly}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                        />
                        <FieldError errors={[engineError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Version"
                        content="Engine or runtime version tag (e.g. 16, 7.2, 20-alpine, latest)."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <Input
                            {...version}
                            value={version.value}
                            onChange={version.onChange}
                            placeholder="e.g. 16, 7.2, latest"
                            aria-invalid={isVersionInvalid}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                            disabled={readOnly}
                        />
                        <FieldError errors={[versionError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Container Port"
                        content="The primary network port exposed by the application container. Changing this will automatically synchronize with the app's routing settings."
                    />
                }
            >
                <FieldGroup>
                    <Field>
                        <InputNumber
                            name={port.name}
                            ref={port.ref}
                            onBlur={port.onBlur}
                            disabled={readOnly || port.disabled}
                            value={port.value ? port.value : undefined}
                            onValueChange={val => {
                                if (readOnly) return;
                                port.onChange(val ?? 0);
                            }}
                            useGrouping={false}
                            placeholder="e.g. 80, 8080, 5432"
                            className="max-w-[140px]"
                            aria-invalid={isPortInvalid}
                        />
                        <FieldError errors={[portError]} />
                    </Field>
                </FieldGroup>
            </InfoBlock>
        </>
    );
}
