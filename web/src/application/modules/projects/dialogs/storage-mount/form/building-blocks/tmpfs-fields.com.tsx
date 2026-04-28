import React from "react";

import { Input } from "@components/ui";
import { Field, FieldError } from "@components/ui/field";
import { useController, useFormContext } from "react-hook-form";
import type { ProjectStorageSettings } from "~/projects/domain";

import { InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";

import type { StorageMountFormInput, StorageMountFormOutput } from "../../schemas";

interface TmpfsFieldsProps {
    projectRules?: ProjectStorageSettings;
}

export function TmpfsFields({ projectRules }: TmpfsFieldsProps) {
    const { control } = useFormContext<StorageMountFormInput, unknown, StorageMountFormOutput>();

    const {
        field: sizeField,
        fieldState: { error: sizeError },
    } = useController({ name: "tmpfsOptions.size", control });
    const { field: modeField } = useController({ name: "tmpfsOptions.mode", control });

    const maxSize = projectRules?.tmpfsSettings?.maxSize;

    return (
        <>
            <Field>
                <InfoBlock title={<LabelWithInfo label="Size (bytes)" />}>
                    <Input
                        {...sizeField}
                        id="size"
                        type="number"
                        placeholder="0"
                        value={sizeField.value ?? ""}
                        onChange={e => {
                            const { value } = e.target;
                            sizeField.onChange(value);
                        }}
                    />
                    {maxSize && <p className="text-xs text-muted-foreground mt-1">Max size: {maxSize} bytes</p>}
                    <FieldError errors={[sizeError]} />
                </InfoBlock>
            </Field>

            <Field>
                <InfoBlock title={<LabelWithInfo label="File Mode (octal)" />}>
                    <Input
                        {...modeField}
                        id="mode"
                        placeholder="0755"
                        value={modeField.value ?? ""}
                        onChange={e => {
                            const { value } = e.target;
                            if (value) {
                                modeField.onChange(parseInt(value, 8));
                            } else {
                                modeField.onChange(undefined);
                            }
                        }}
                    />
                </InfoBlock>
            </Field>
        </>
    );
}
