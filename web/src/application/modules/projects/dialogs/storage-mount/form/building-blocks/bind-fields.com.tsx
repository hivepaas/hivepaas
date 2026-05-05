import React, { useEffect, useMemo } from "react";

import { cn } from "@/lib/utils";
import { Input } from "@components/ui";
import { Field, FieldError } from "@components/ui/field";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { dashedBorderBox } from "@lib/styles";
import { useController, useFormContext, useWatch } from "react-hook-form";
import type { ProjectStorageSettings } from "~/projects/domain";
import { EMountPropagation } from "~/projects/module-shared/enums";

import { EditableCombobox, InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { StorageMountFormInput, StorageMountFormOutput } from "../../schemas";
import { computeRequiredSubpath } from "../../utils/required-subpath.util";

interface BindFieldsProps {
    projectRules?: ProjectStorageSettings;
    projectKey?: string;
    appLocalKey?: string;
}

export function BindFields({ projectRules, projectKey, appLocalKey }: BindFieldsProps) {
    const { control, setValue } = useFormContext<StorageMountFormInput, unknown, StorageMountFormOutput>();

    const {
        field: baseDirField,
        fieldState: { invalid: baseDirInvalid, error: baseDirError },
    } = useController({ name: "bindOptions.baseDir", control });
    const { field: subpathField } = useController({ name: "bindOptions.subpath", control });
    const { field: propagationField } = useController({ name: "bindOptions.propagation", control });

    const baseDirs = projectRules?.bindSettings?.baseDirs ?? [];

    const requiredPrefix = useMemo(
        () =>
            computeRequiredSubpath(
                {
                    enabled: projectRules?.bindSettings?.enabled,
                    baseSubpath: projectRules?.bindSettings?.baseSubpath,
                    appsMustUseSubPaths: projectRules?.bindSettings?.appsMustUseSubPaths,
                },
                projectKey,
                appLocalKey,
            ),
        [projectRules?.bindSettings, projectKey, appLocalKey],
    );

    const subpathValue = useWatch({ control, name: "bindOptions.subpath" }) ?? "";

    useEffect(() => {
        setValue("bindOptions.subpathRequired", requiredPrefix, { shouldDirty: false, shouldValidate: false });
        if (requiredPrefix && !subpathValue) {
            // Auto-fill prefix once as initial value, then let user edit freely.
            setValue("bindOptions.subpath", requiredPrefix, { shouldDirty: false, shouldValidate: false });
        }
    }, [requiredPrefix, setValue, subpathValue]);

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
                    />
                    <FieldError errors={[baseDirError]} />
                </InfoBlock>
            </Field>

            {requiredPrefix !== "" && (
                <Field>
                    <InfoBlock
                        title={<LabelWithInfo label="Required Subpath Prefix" />}
                        titleWidth={180}
                    >
                        <div className={cn(dashedBorderBox, "py-2 px-3 rounded-md")}>{requiredPrefix}</div>
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
                        placeholder={requiredPrefix}
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
