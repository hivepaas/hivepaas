import React from "react";

import { Input } from "@components/ui";
import { Field, FieldError } from "@components/ui/field";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { useController, useFormContext } from "react-hook-form";
import type { ProjectStorageSettings } from "~/projects/domain";
import { EMountPropagation } from "~/projects/module-shared/enums";

import { EditableCombobox, InfoBlock, LabelWithInfo } from "@application/shared/components";

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
                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="Base Source Dir"
                            isRequired
                        />
                    }
                >
                    <EditableCombobox
                        options={baseDirs}
                        value={baseDirField.value || ""}
                        onChange={baseDirField.onChange}
                        placeholder="/data"
                        aria-invalid={baseDirInvalid}
                        emptyText={baseDirs.length === 0 ? "No base directories available" : "No matching directories"}
                    />
                    <FieldError errors={[baseDirError]} />
                </InfoBlock>
            </Field>

            <Field>
                <InfoBlock title={<LabelWithInfo label="Required Subpath Prefix" />}>
                    {projectRules?.bindSettings?.appsMustUseSubPaths && (
                        <p className="text-xs text-muted-foreground mt-1">{projectRules.bindSettings.baseSubpath}</p>
                    )}
                </InfoBlock>
            </Field>

            <Field>
                <InfoBlock title={<LabelWithInfo label="Subpath" />}>
                    <Input
                        {...subpathField}
                        id="subpath"
                        placeholder="app/data"
                    />
                </InfoBlock>
            </Field>

            <Field>
                <InfoBlock title={<LabelWithInfo label="Propagation" />}>
                    <Select
                        {...propagationField}
                        value={propagationField.value ?? EMountPropagation.RPrivate}
                        onValueChange={propagationField.onChange}
                    >
                        <SelectTrigger>
                            <SelectValue placeholder="Propagation" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value={EMountPropagation.RPrivate}>rprivate</SelectItem>
                            <SelectItem value={EMountPropagation.Private}>private</SelectItem>
                            <SelectItem value={EMountPropagation.RShared}>rshared</SelectItem>
                            <SelectItem value={EMountPropagation.Shared}>shared</SelectItem>
                            <SelectItem value={EMountPropagation.RSlave}>rslave</SelectItem>
                            <SelectItem value={EMountPropagation.Slave}>slave</SelectItem>
                        </SelectContent>
                    </Select>
                </InfoBlock>
            </Field>
        </>
    );
}
