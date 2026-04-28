import React from "react";

import { Field, FieldError, FieldLabel } from "@components/ui/field";
import { SelectItem } from "@components/ui/select";
import { useController, useFormContext } from "react-hook-form";
import type { ProjectStorageSettings } from "~/projects/domain";
import { EMountPropagation } from "~/projects/module-shared/enums";

import { InputWithAddOn, SelectWithAddon } from "@application/shared/components";

import type { StorageMountFormInput, StorageMountFormOutput } from "../../schemas";

interface BindFieldsProps {
    projectRules?: ProjectStorageSettings;
}

export function BindFields({ projectRules }: BindFieldsProps) {
    const { control } = useFormContext<StorageMountFormInput, unknown, StorageMountFormOutput>();

    const {
        field: baseDirField,
        fieldState: { invalid: baseDirInvalid, error: baseDirError },
    } = useController({ name: "bindOptions.baseDir", control });
    const { field: subpathField } = useController({ name: "bindOptions.subpath", control });
    const { field: propagationField } = useController({ name: "bindOptions.propagation", control });

    const baseDirs = projectRules?.bindSettings?.baseDirs ?? [];

    return (
        <>
            <Field>
                <FieldLabel htmlFor="baseDir">Base Directory *</FieldLabel>
                {baseDirs.length > 0 ? (
                    <SelectWithAddon
                        {...baseDirField}
                        addonLeft="Base Dir"
                        value={baseDirField.value || ""}
                        onValueChange={baseDirField.onChange}
                    >
                        {baseDirs.map((dir: string) => (
                            <SelectItem
                                key={dir}
                                value={dir}
                            >
                                {dir}
                            </SelectItem>
                        ))}
                    </SelectWithAddon>
                ) : (
                    <InputWithAddOn
                        {...baseDirField}
                        id="baseDir"
                        placeholder="/data"
                        aria-invalid={baseDirInvalid}
                        addonLeft="Base Dir"
                    />
                )}
                <FieldError errors={[baseDirError]} />
            </Field>

            <Field>
                <FieldLabel htmlFor="subpath">Subpath</FieldLabel>
                <InputWithAddOn
                    {...subpathField}
                    id="subpath"
                    placeholder="app/data"
                    addonLeft="Subpath"
                />
                {projectRules?.bindSettings?.appsMustUseSubPaths && (
                    <p className="text-xs text-muted-foreground mt-1">
                        Required prefix: {projectRules.bindSettings.baseSubpath}
                    </p>
                )}
            </Field>

            <Field>
                <FieldLabel htmlFor="propagation">Propagation</FieldLabel>
                <SelectWithAddon
                    {...propagationField}
                    addonLeft="Propagation"
                    value={propagationField.value ?? EMountPropagation.RPrivate}
                    onValueChange={propagationField.onChange}
                >
                    <SelectItem value={EMountPropagation.RPrivate}>rprivate</SelectItem>
                    <SelectItem value={EMountPropagation.Private}>private</SelectItem>
                    <SelectItem value={EMountPropagation.RShared}>rshared</SelectItem>
                    <SelectItem value={EMountPropagation.Shared}>shared</SelectItem>
                    <SelectItem value={EMountPropagation.RSlave}>rslave</SelectItem>
                    <SelectItem value={EMountPropagation.Slave}>slave</SelectItem>
                </SelectWithAddon>
            </Field>
        </>
    );
}
