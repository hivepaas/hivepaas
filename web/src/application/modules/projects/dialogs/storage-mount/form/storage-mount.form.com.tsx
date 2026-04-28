import React, { useEffect } from "react";

import { Button } from "@components/ui/button";
import { Field, FieldError, FieldGroup, FieldLabel } from "@components/ui/field";
import { SelectItem } from "@components/ui/select";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useController, useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";
import type { ProjectStorageSettings } from "~/projects/domain";
import { EMountConsistency, EMountType } from "~/projects/module-shared/enums";

import { InputWithAddOn, SelectWithAddon } from "@application/shared/components";

import type { StorageMountFormInput, StorageMountFormOutput } from "../schemas";
import { StorageMountFormSchema } from "../schemas";

import { BindFields, ClusterFields, TmpfsFields, VolumeFields } from "./building-blocks";

export function StorageMountForm({ isPending, onSubmit, initialValues, projectRules }: Props) {
    const methods = useForm<StorageMountFormInput, unknown, StorageMountFormOutput>({
        resolver: zodResolver(StorageMountFormSchema),
        mode: "onSubmit",
        defaultValues: initialValues ?? {
            type: EMountType.Bind,
            target: "",
            readOnly: false,
            consistency: EMountConsistency.Default,
            bindOptions: {
                baseDir: "",
                subpath: "",
            },
        },
    });

    const {
        handleSubmit,
        control,
        reset,
        formState: { errors },
    } = methods;

    const { field: typeField } = useController({ name: "type", control });
    const {
        field: targetField,
        fieldState: { invalid: targetInvalid },
    } = useController({ name: "target", control });
    const { field: readOnlyField } = useController({ name: "readOnly", control });
    const { field: consistencyField } = useController({ name: "consistency", control });

    const storageType = useWatch({ control, name: "type" });

    useEffect(() => {
        if (initialValues) {
            reset(initialValues);
        }
    }, [initialValues, reset]);

    const availableTypes = React.useMemo(() => {
        const types: EMountType[] = [];
        if (projectRules?.bindSettings?.enabled) types.push(EMountType.Bind);
        if (projectRules?.volumeSettings?.enabled) types.push(EMountType.Volume);
        if (projectRules?.clusterVolumeSettings?.enabled) types.push(EMountType.Cluster);
        if (projectRules?.tmpfsSettings?.enabled) types.push(EMountType.Tmpfs);
        return types.length > 0 ? types : [EMountType.Bind, EMountType.Volume, EMountType.Cluster, EMountType.Tmpfs];
    }, [projectRules]);

    function onValid(data: StorageMountFormOutput) {
        void onSubmit(data);
    }

    function onInvalid(_errors: FieldErrors<StorageMountFormOutput>) {
        toast.error("Please fix the validation errors");
    }

    return (
        <FormProvider {...methods}>
            <form
                onSubmit={e => {
                    e.preventDefault();
                    void handleSubmit(onValid, onInvalid)(e);
                }}
            >
                <FieldGroup>
                    <Field>
                        <FieldLabel htmlFor="type">Storage Type *</FieldLabel>
                        <SelectWithAddon
                            {...typeField}
                            addonLeft="Type"
                            value={typeField.value}
                            onValueChange={typeField.onChange}
                        >
                            {availableTypes.map(type => (
                                <SelectItem
                                    key={type}
                                    value={type}
                                >
                                    {type}
                                </SelectItem>
                            ))}
                        </SelectWithAddon>
                        <FieldError errors={[errors.type]} />
                    </Field>

                    {storageType === EMountType.Bind && <BindFields projectRules={projectRules} />}

                    {storageType === EMountType.Volume && <VolumeFields projectRules={projectRules} />}

                    {storageType === EMountType.Cluster && <ClusterFields projectRules={projectRules} />}

                    {storageType === EMountType.Tmpfs && <TmpfsFields projectRules={projectRules} />}

                    <Field>
                        <FieldLabel htmlFor="target">Target *</FieldLabel>
                        <InputWithAddOn
                            {...targetField}
                            id="target"
                            placeholder="/path/in/container"
                            aria-invalid={targetInvalid}
                            addonLeft="Target"
                        />
                        <FieldError errors={[errors.target]} />
                    </Field>

                    <Field>
                        <label className="flex items-center gap-2">
                            <input
                                type="checkbox"
                                checked={readOnlyField.value ?? false}
                                onChange={readOnlyField.onChange}
                            />
                            <span className="text-sm">Read-only</span>
                        </label>
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="consistency">Consistency</FieldLabel>
                        <SelectWithAddon
                            {...consistencyField}
                            addonLeft="Consistency"
                            value={consistencyField.value ?? EMountConsistency.Default}
                            onValueChange={consistencyField.onChange}
                        >
                            <SelectItem value={EMountConsistency.Default}>default</SelectItem>
                            <SelectItem value={EMountConsistency.Consistent}>consistent</SelectItem>
                            <SelectItem value={EMountConsistency.Cached}>cached</SelectItem>
                            <SelectItem value={EMountConsistency.Delegated}>delegated</SelectItem>
                        </SelectWithAddon>
                    </Field>

                    <div className="flex justify-end gap-2 pt-4">
                        <Button
                            type="submit"
                            isLoading={isPending}
                        >
                            {initialValues ? "Update" : "Add"}
                        </Button>
                    </div>
                </FieldGroup>
            </form>
        </FormProvider>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: StorageMountFormOutput) => Promise<void> | void;
    initialValues?: StorageMountFormInput;
    projectRules?: ProjectStorageSettings;
}
