import React, { useImperativeHandle } from "react";

import { Checkbox, Input } from "@components/ui";
import { Button } from "@components/ui/button";
import { Field, FieldError, FieldGroup } from "@components/ui/field";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, type FieldPath, FormProvider, useController, useForm, useWatch } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { toast } from "sonner";
import type { AppStorageMount, ProjectStorageSettings } from "~/projects/domain";
import { EMountConsistency, EMountType } from "~/projects/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { ValidationException } from "@infrastructure/exceptions/validation";

import type { StorageMountFormInput, StorageMountFormOutput } from "../schemas";
import { StorageMountFormSchema } from "../schemas";
import type { StorageMountFormRef } from "../types";

import { BindFields, ClusterFields, TmpfsFields, VolumeFields } from "./building-blocks";
import { emptyStorageMountFormDefaults, mountToFormInput } from "./storage-mount.form-mappers";

type Props = {
    ref?: React.Ref<StorageMountFormRef>;
    isPending: boolean;
    defaultValues?: AppStorageMount;
    onSubmit: (values: StorageMountFormOutput) => void;
    projectRules?: ProjectStorageSettings;
    projectKey?: string;
    appLocalKey?: string;
};

export function StorageMountForm({
    ref,
    isPending,
    onSubmit,
    defaultValues,
    projectRules,
    projectKey,
    appLocalKey,
}: Props) {
    const methods = useForm<StorageMountFormInput, unknown, StorageMountFormOutput>({
        defaultValues: defaultValues ? mountToFormInput(defaultValues) : emptyStorageMountFormDefaults,
        resolver: zodResolver(StorageMountFormSchema),
        mode: "onSubmit",
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

    useUpdateEffect(() => {
        reset(defaultValues ? mountToFormInput(defaultValues) : emptyStorageMountFormDefaults);
    }, [defaultValues, reset]);

    useImperativeHandle(
        ref,
        () => ({
            setValues: (values: Partial<StorageMountFormInput>) => {
                reset({
                    ...methods.getValues(),
                    ...values,
                } as StorageMountFormInput);
            },
            onError(error: ValidationException) {
                if (error.errors.length === 0) {
                    return;
                }
                error.errors.forEach(({ path, message }, index) => {
                    methods.setError(
                        path as FieldPath<StorageMountFormInput>,
                        { message, type: "manual" },
                        { shouldFocus: index === 0 },
                    );
                });
            },
        }),
        [methods, reset],
    );

    const availableTypes = React.useMemo(() => {
        const types: EMountType[] = [];
        if (projectRules?.bindSettings?.enabled) types.push(EMountType.Bind);
        if (projectRules?.volumeSettings?.enabled) types.push(EMountType.Volume);
        if (projectRules?.clusterVolumeSettings?.enabled) types.push(EMountType.Cluster);
        if (projectRules?.tmpfsSettings?.enabled) types.push(EMountType.Tmpfs);
        return types.length > 0 ? types : [];
    }, [projectRules]);

    function onValid(data: StorageMountFormOutput) {
        onSubmit(data);
    }

    function onInvalid(_errors: FieldErrors<StorageMountFormInput>) {
        console.error(_errors);
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
                <FieldGroup className="gap-6">
                    <Field>
                        <InfoBlock
                            title={
                                <LabelWithInfo
                                    label="Type"
                                    isRequired
                                />
                            }
                            titleWidth={180}
                        >
                            <Select
                                {...typeField}
                                value={typeField.value}
                                onValueChange={typeField.onChange}
                            >
                                <SelectTrigger className="w-[220px]">
                                    <SelectValue placeholder="Type" />
                                </SelectTrigger>
                                <SelectContent>
                                    {availableTypes.map(type => (
                                        <SelectItem
                                            key={type}
                                            value={type}
                                        >
                                            {type}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </InfoBlock>
                        <FieldError errors={[errors.type]} />
                    </Field>

                    {storageType === EMountType.Bind && (
                        <BindFields
                            projectRules={projectRules}
                            projectKey={projectKey}
                            appLocalKey={appLocalKey}
                        />
                    )}

                    {storageType === EMountType.Volume && (
                        <VolumeFields
                            projectRules={projectRules}
                            projectKey={projectKey}
                            appLocalKey={appLocalKey}
                        />
                    )}

                    {storageType === EMountType.Cluster && (
                        <ClusterFields
                            projectRules={projectRules}
                            projectKey={projectKey}
                            appLocalKey={appLocalKey}
                        />
                    )}

                    {storageType === EMountType.Tmpfs && <TmpfsFields projectRules={projectRules} />}

                    {storageType !== EMountType.Tmpfs && (
                        <Field>
                            <InfoBlock
                                title={<LabelWithInfo label="Read-only" />}
                                titleWidth={180}
                            >
                                <Checkbox
                                    id="read-only"
                                    checked={readOnlyField.value ?? false}
                                    onCheckedChange={checked => {
                                        readOnlyField.onChange(checked === true);
                                    }}
                                />
                            </InfoBlock>
                        </Field>
                    )}

                    <Field>
                        <InfoBlock
                            title={
                                <LabelWithInfo
                                    label="Target"
                                    isRequired
                                />
                            }
                            titleWidth={180}
                        >
                            <Input
                                {...targetField}
                                id="target"
                                placeholder="/path/in/container"
                                aria-invalid={targetInvalid}
                            />
                            <FieldError errors={[errors.target]} />
                        </InfoBlock>
                    </Field>

                    <Field>
                        <InfoBlock
                            title={<LabelWithInfo label="Consistency" />}
                            titleWidth={180}
                        >
                            <Select
                                {...consistencyField}
                                value={consistencyField.value ?? EMountConsistency.Default}
                                onValueChange={consistencyField.onChange}
                            >
                                <SelectTrigger className="w-[220px]">
                                    <SelectValue placeholder="Consistency" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value={EMountConsistency.Default}>default</SelectItem>
                                    <SelectItem value={EMountConsistency.Consistent}>consistent</SelectItem>
                                    <SelectItem value={EMountConsistency.Cached}>cached</SelectItem>
                                    <SelectItem value={EMountConsistency.Delegated}>delegated</SelectItem>
                                </SelectContent>
                            </Select>
                        </InfoBlock>
                    </Field>

                    <div className="flex justify-end gap-2 pt-4">
                        <Button
                            type="submit"
                            isLoading={isPending}
                        >
                            {defaultValues ? "Update" : "Add"}
                        </Button>
                    </div>
                </FieldGroup>
            </form>
        </FormProvider>
    );
}
