import React, { useImperativeHandle } from "react";

import { Checkbox, Input } from "@components/ui";
import { Button } from "@components/ui/button";
import { Field, FieldError, FieldGroup } from "@components/ui/field";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, type FieldPath, FormProvider, useController, useForm, useWatch } from "react-hook-form";
import { useParams } from "react-router";
import { useUpdateEffect } from "react-use";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppStorageSettingsQueries } from "~/projects/data";
import type { AppStorageMount } from "~/projects/domain";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { EMountConsistency, EMountType } from "~/projects/module-shared/enums";

import {
    AppLink,
    Combobox,
    type ComboboxOption,
    FormActionBar,
    InfoBlock,
    LabelWithInfo,
} from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

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
    projectKey?: string;
    appKey?: string;
    readOnly?: boolean;
    stickyActions?: boolean;
    onClose?: () => void;
    children?: React.ReactNode;
};

export function StorageMountForm({
    ref,
    isPending,
    onSubmit,
    defaultValues,
    readOnly = false,
    stickyActions = false,
    onClose,
    children,
}: Props) {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const {
        data: storageSettingsData,
        isFetching: isFetchingStorageSettings,
        refetch: refetchStorageSettings,
        isRefetching: isRefetchingStorageSettings,
    } = AppStorageSettingsQueries.useFindOne({
        projectID: projectId,
        appID: appId,
    });

    const storageSettings = storageSettingsData?.data.settings;

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

    const typeOptions = React.useMemo<ComboboxOption<{ id: EMountType; name: string }>[]>(() => {
        const options: ComboboxOption<{ id: EMountType; name: string }>[] = [];
        if (storageSettings?.bindSettings?.enabled) {
            options.push({ value: { id: EMountType.Bind, name: EMountType.Bind }, label: EMountType.Bind });
        }
        if (storageSettings?.volumeSettings?.enabled) {
            options.push({ value: { id: EMountType.Volume, name: EMountType.Volume }, label: EMountType.Volume });
        }
        if (storageSettings?.clusterVolumeSettings?.enabled) {
            options.push({ value: { id: EMountType.Cluster, name: EMountType.Cluster }, label: EMountType.Cluster });
        }
        if (storageSettings?.tmpfsSettings?.enabled) {
            options.push({ value: { id: EMountType.Tmpfs, name: EMountType.Tmpfs }, label: EMountType.Tmpfs });
        }
        return options;
    }, [storageSettings]);

    function onValid(data: StorageMountFormOutput) {
        if (readOnly) {
            return;
        }

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
                    if (readOnly) {
                        return;
                    }

                    void handleSubmit(onValid, onInvalid)(e);
                }}
                className="min-h-0 flex flex-1 flex-col"
            >
                <fieldset
                    disabled={readOnly}
                    className="contents"
                >
                    <div className="flex flex-col gap-6">
                        {children}
                        <FieldGroup className="gap-6">
                            <InfoBlock
                                title={
                                    <LabelWithInfo
                                        label="Type"
                                        isRequired
                                    />
                                }
                                titleWidth={180}
                            >
                                <Field>
                                    <Combobox
                                        options={typeOptions}
                                        value={typeField.value}
                                        onChange={value => {
                                            typeField.onChange(value ?? undefined);
                                        }}
                                        placeholder="Type"
                                        searchable={false}
                                        closeOnSelect
                                        emptyText="No storage types available"
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                        valueKey="id"
                                        aria-invalid={Boolean(errors.type)}
                                        loading={isFetchingStorageSettings}
                                        onRefresh={() => void refetchStorageSettings()}
                                        isRefreshing={isRefetchingStorageSettings}
                                    />
                                    <FieldError errors={[errors.type]} />
                                    <div className="text-xs">
                                        <p>
                                            <AppLink.Basic
                                                to={ROUTE.projects.single.configuration.storageSettings.$route(
                                                    projectId,
                                                )}
                                                className="text-link"
                                                target="_blank"
                                                rel="noopener noreferrer"
                                            >
                                                Configure storage settings in the project
                                            </AppLink.Basic>
                                        </p>
                                    </div>
                                </Field>
                            </InfoBlock>

                            {storageType === EMountType.Bind && <BindFields storageSettings={storageSettings} />}

                            {storageType === EMountType.Volume && <VolumeFields storageSettings={storageSettings} />}

                            {storageType === EMountType.Cluster && <ClusterFields storageSettings={storageSettings} />}

                            {storageType === EMountType.Tmpfs && <TmpfsFields storageSettings={storageSettings} />}

                            {storageType !== EMountType.Tmpfs && (
                                <Field>
                                    <InfoBlock
                                        title={<LabelWithInfo label="Read Only" />}
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
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
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
                        </FieldGroup>
                    </div>
                    {!readOnly && (
                        <FormActionBar sticky={stickyActions}>
                            <Button
                                type="button"
                                variant="outline"
                                className="min-w-[100px]"
                                disabled={isPending}
                                onClick={onClose}
                            >
                                Cancel
                            </Button>
                            <Button
                                type="submit"
                                isLoading={isPending}
                            >
                                {defaultValues ? "Update" : "Add"}
                            </Button>
                        </FormActionBar>
                    )}
                    {readOnly && (
                        <FormActionBar sticky={stickyActions}>
                            <Button
                                type="button"
                                onClick={onClose}
                                className="min-w-[100px]"
                            >
                                Close
                            </Button>
                        </FormActionBar>
                    )}
                </fieldset>
            </form>
        </FormProvider>
    );
}
