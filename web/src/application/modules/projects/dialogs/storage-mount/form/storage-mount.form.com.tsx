import React, { useMemo, useState } from "react";

import { Checkbox, Input } from "@components/ui";
import { Button } from "@components/ui/button";
import { Field, FieldError, FieldGroup } from "@components/ui/field";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { zodResolver } from "@hookform/resolvers/zod";
import { FormProvider, useController, useForm } from "react-hook-form";
import { useUpdateEffect } from "react-use";
import { ProjectClusterVolumesQueries } from "~/projects/data/queries";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import { EMountConsistency } from "~/projects/module-shared/enums";

import { AppLink, Combobox, FormActionBar, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";

import type { StorageMountFormInput, StorageMountFormOutput } from "../schemas";
import { StorageMountFormSchema } from "../schemas";

import { emptyStorageMountFormDefaults } from "./storage-mount.form-mappers";

type Props = {
    projectId: string;
    isPending: boolean;
    isEditMode?: boolean;
    defaultValues?: StorageMountFormInput;
    onSubmit: (values: StorageMountFormOutput) => void;
    readOnly?: boolean;
    onClose?: () => void;
    children?: React.ReactNode;
};

export function StorageMountForm({
    projectId,
    isPending,
    isEditMode = false,
    defaultValues,
    onSubmit,
    readOnly = false,
    onClose,
    children,
}: Props) {
    const [searchQuery, setSearchQuery] = useState("");

    const {
        data: { data: volumes } = DEFAULT_PAGINATED_DATA,
        isFetching,
        refetch,
        isRefetching,
    } = ProjectClusterVolumesQueries.useFindManyPaginated({
        projectID: projectId,
        search: searchQuery,
    });

    const methods = useForm<StorageMountFormInput, unknown, StorageMountFormOutput>({
        defaultValues: defaultValues ?? emptyStorageMountFormDefaults,
        resolver: zodResolver(StorageMountFormSchema),
        mode: "onSubmit",
    });

    const { handleSubmit, control, reset } = methods;

    useUpdateEffect(() => {
        reset(defaultValues ?? emptyStorageMountFormDefaults);
    }, [defaultValues, reset]);

    const {
        field: sourceField,
        fieldState: { invalid: sourceInvalid, error: sourceError },
    } = useController({ name: "source", control });
    const {
        field: subpathField,
        fieldState: { invalid: subpathInvalid, error: subpathError },
    } = useController({ name: "subpath", control });
    const { field: readOnlyField } = useController({ name: "readOnly", control });
    const { field: noCopyField } = useController({ name: "noCopy", control });
    const {
        field: targetField,
        fieldState: { invalid: targetInvalid, error: targetError },
    } = useController({ name: "target", control });
    const { field: consistencyField } = useController({ name: "consistency", control });

    const volumeOptions = useMemo(() => {
        return volumes.map(volume => ({
            value: { id: volume.id },
            label: volume.name,
        }));
    }, [volumes]);

    return (
        <FormProvider {...methods}>
            <form
                onSubmit={event => {
                    void handleSubmit(onSubmit)(event);
                }}
                className="min-h-0 flex flex-1 flex-col"
            >
                <fieldset
                    disabled={readOnly}
                    className="contents"
                >
                    {children}

                    <FieldGroup>
                        <Field>
                            <InfoBlock
                                title={
                                    <LabelWithInfo
                                        label="Volume"
                                        isRequired
                                    />
                                }
                                titleWidth={180}
                            >
                                <Combobox
                                    options={volumeOptions}
                                    value={sourceField.value || null}
                                    onChange={value => {
                                        sourceField.onChange(value ?? "");
                                    }}
                                    onSearch={setSearchQuery}
                                    placeholder="select volume"
                                    searchable
                                    closeOnSelect
                                    emptyText="No volumes available"
                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    valueKey="id"
                                    aria-invalid={sourceInvalid}
                                    loading={isFetching}
                                    onRefresh={() => void refetch()}
                                    isRefreshing={isRefetching}
                                    disabled={readOnly}
                                />
                                <FieldError errors={[sourceError]} />
                                <div className="text-xs">
                                    Need to add new volumes?{" "}
                                    <AppLink.Basic
                                        to={ROUTE.projects.single.clusterResources.volumes.$route(projectId)}
                                        className="text-link"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                    >
                                        Click here
                                    </AppLink.Basic>
                                </div>
                            </InfoBlock>
                        </Field>

                        <Field>
                            <InfoBlock
                                title={<LabelWithInfo label="Subpath" />}
                                titleWidth={180}
                            >
                                <Input
                                    {...subpathField}
                                    id="subpath"
                                    placeholder="subpath"
                                    aria-invalid={subpathInvalid}
                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                />
                                <FieldError errors={[subpathError]} />
                            </InfoBlock>
                        </Field>

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

                        <Field>
                            <InfoBlock
                                title={<LabelWithInfo label="No Copy" />}
                                titleWidth={180}
                            >
                                <Checkbox
                                    id="no-copy"
                                    checked={noCopyField.value ?? false}
                                    onCheckedChange={checked => {
                                        noCopyField.onChange(checked === true);
                                    }}
                                />
                            </InfoBlock>
                        </Field>

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
                                <FieldError errors={[targetError]} />
                            </InfoBlock>
                        </Field>

                        <Field>
                            <InfoBlock
                                title={<LabelWithInfo label="Consistency" />}
                                titleWidth={180}
                            >
                                <Select
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

                    {!readOnly && (
                        <FormActionBar>
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
                                {isEditMode ? "Update" : "Save"}
                            </Button>
                        </FormActionBar>
                    )}
                </fieldset>
            </form>
        </FormProvider>
    );
}
