import React from "react";

import { Input } from "@components/ui";
import { Field, FieldError } from "@components/ui/field";
import { useController, useFormContext } from "react-hook-form";
import type { SettingStorageSettings } from "~/settings/domain";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { StorageMountFormInput, StorageMountFormOutput } from "../../schemas";

interface TmpfsFieldsProps {
    storageSettings?: SettingStorageSettings;
}

export function TmpfsFields({ storageSettings }: TmpfsFieldsProps) {
    const { control } = useFormContext<StorageMountFormInput, unknown, StorageMountFormOutput>();

    const {
        field: sizeField,
        fieldState: { error: sizeError },
    } = useController({ name: "tmpfsOptions.size", control });
    const { field: modeField } = useController({ name: "tmpfsOptions.mode", control });

    const maxSize = storageSettings?.tmpfsSettings?.maxSize;

    return (
        <>
            <Field>
                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="Size"
                            isRequired
                        />
                    }
                    titleWidth={180}
                >
                    <Input
                        {...sizeField}
                        id="size"
                        placeholder="0"
                        value={sizeField.value}
                        onChange={e => {
                            const { value } = e.target;
                            sizeField.onChange(value);
                        }}
                        className="w-[220px]"
                    />
                    {maxSize && <p className="text-xs text-muted-foreground mt-1">Max size: {maxSize}</p>}
                    <FieldError errors={[sizeError]} />
                </InfoBlock>
            </Field>

            <Field>
                <InfoBlock
                    title={<LabelWithInfo label="File Mode" />}
                    titleWidth={180}
                >
                    <Input
                        {...modeField}
                        id="mode"
                        placeholder="0755"
                        value={modeField.value}
                        onChange={e => {
                            const { value } = e.target;
                            if (value) {
                                modeField.onChange(value);
                            } else {
                                modeField.onChange(undefined);
                            }
                        }}
                        className="w-[220px]"
                    />
                </InfoBlock>
            </Field>
        </>
    );
}
