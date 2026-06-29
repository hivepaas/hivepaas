import React from "react";

import { cn } from "@/lib/utils";
import { Input } from "@components/ui";
import { Field, FieldError } from "@components/ui/field";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { dashedBorderBox } from "@lib/styles";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { EMountPropagation } from "~/projects/module-shared/enums";
import type { SettingStorageSettings } from "~/settings/domain";

import { EditableCombobox, InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { StorageMountFormInput, StorageMountFormOutput } from "../../schemas";

interface BindFieldsProps {
    storageSettings?: SettingStorageSettings;
    projectKey?: string;
    appLocalKey?: string;
}

export function BindFields({ storageSettings }: BindFieldsProps) {
    const { control } = useFormContext<StorageMountFormInput, unknown, StorageMountFormOutput>();

    const {
        field: baseDirField,
        fieldState: { invalid: baseDirInvalid, error: baseDirError },
    } = useController({ name: "bindOptions.baseDir", control });
    const { field: subpathField } = useController({ name: "bindOptions.subpath", control });
    const { field: propagationField } = useController({ name: "bindOptions.propagation", control });
    const { field: subpathRequiredField } = useController({ name: "bindOptions.subpathRequired", control });

    const baseDirs = storageSettings?.bindSettings?.baseDirs ?? [];

    const subpathValue = useWatch({ control, name: "bindOptions.subpath" }) ?? "";

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
                    titleWidth={180}
                >
                    <EditableCombobox
                        options={baseDirs}
                        value={baseDirField.value || ""}
                        onChange={baseDirField.onChange}
                        placeholder="/data"
                        aria-invalid={baseDirInvalid}
                        emptyText={baseDirs.length === 0 ? "No base directories available" : "No matching directories"}
                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                    />
                    <FieldError errors={[baseDirError]} />
                </InfoBlock>
            </Field>

            {subpathRequiredField.value && (
                <Field>
                    <InfoBlock
                        title={<LabelWithInfo label="Required Subpath Prefix" />}
                        titleWidth={180}
                    >
                        <div
                            className={cn(
                                dashedBorderBox,
                                PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS,
                                "py-2 px-3 rounded-md",
                            )}
                        >
                            {subpathRequiredField.value}
                        </div>
                    </InfoBlock>
                </Field>
            )}

            <Field>
                <InfoBlock
                    title={<LabelWithInfo label="Subpath" />}
                    titleWidth={180}
                >
                    <Input
                        {...subpathField}
                        id="subpath"
                        value={subpathValue}
                        placeholder={subpathRequiredField.value}
                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                    />
                </InfoBlock>
            </Field>

            <Field>
                <InfoBlock
                    title={<LabelWithInfo label="Propagation" />}
                    titleWidth={180}
                >
                    <Select
                        {...propagationField}
                        value={propagationField.value ?? EMountPropagation.RPrivate}
                        onValueChange={propagationField.onChange}
                    >
                        <SelectTrigger className="w-[220px]">
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
