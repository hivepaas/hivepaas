import React from "react";

import { Field, FieldError, FieldLabel } from "@components/ui/field";
import { useController, useFormContext } from "react-hook-form";
import type { ProjectStorageSettings } from "~/projects/domain";

import { InputWithAddOn } from "@application/shared/components";

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
                <FieldLabel htmlFor="size">Size (bytes)</FieldLabel>
                <InputWithAddOn
                    {...sizeField}
                    id="size"
                    type="number"
                    placeholder="0"
                    addonLeft="Size"
                    value={sizeField.value ?? ""}
                    onChange={e => {
                        const { value } = e.target;
                        sizeField.onChange(value);
                    }}
                />
                {maxSize && <p className="text-xs text-muted-foreground mt-1">Max size: {maxSize} bytes</p>}
                <FieldError errors={[sizeError]} />
            </Field>

            <Field>
                <FieldLabel htmlFor="mode">File Mode (octal)</FieldLabel>
                <InputWithAddOn
                    {...modeField}
                    id="mode"
                    placeholder="0755"
                    addonLeft="Mode"
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
            </Field>
        </>
    );
}
