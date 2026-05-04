import React, { useEffect, useMemo } from "react";

import { cn } from "@/lib/utils";
import { Checkbox, Input } from "@components/ui";
import { Field, FieldError } from "@components/ui/field";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { dashedBorderBox } from "@lib/styles";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { useParams } from "react-router";
import { ProjectDockerVolumesQueries } from "~/projects/data/queries/project-docker-volumes";
import type { ProjectStorageSettings } from "~/projects/domain";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form/key-value-list/key-value-list.com";

import type { StorageMountFormInput, StorageMountFormOutput } from "../../schemas";
import { computeRequiredSubpath } from "../../utils/required-subpath.util";

interface VolumeFieldsProps {
    projectRules?: ProjectStorageSettings;
    projectKey?: string;
    appLocalKey?: string;
}

export function VolumeFields({ projectRules, projectKey, appLocalKey }: VolumeFieldsProps) {
    const { control, setValue } = useFormContext<StorageMountFormInput, unknown, StorageMountFormOutput>();
    const { id: projectId } = useParams<{ id: string }>();

    const {
        field: volumeField,
        fieldState: { invalid: volumeInvalid, error: volumeError },
    } = useController({ name: "volumeOptions.volume", control });
    const { field: subpathField } = useController({ name: "volumeOptions.subpath", control });
    const { field: noCopyField } = useController({ name: "volumeOptions.noCopy", control });

    const { data: volumesData, isLoading: isLoadingVolumes } = ProjectDockerVolumesQueries.useList(
        {
            projectID: projectId ?? "",
        },
        {
            enabled: !!projectId,
        },
    );

    const volumes = volumesData?.data ?? [];

    const requiredPrefix = useMemo(
        () =>
            computeRequiredSubpath(
                {
                    enabled: projectRules?.volumeSettings?.enabled,
                    baseSubpath: projectRules?.volumeSettings?.baseSubpath,
                    appsMustUseSubPaths: projectRules?.volumeSettings?.appsMustUseSubPaths,
                    hasItems: (projectRules?.volumeSettings?.volumes?.length ?? 0) > 0,
                },
                projectKey,
                appLocalKey,
            ),
        [projectRules?.volumeSettings, projectKey, appLocalKey],
    );

    const subpathValue = useWatch({ control, name: "volumeOptions.subpath" }) ?? "";
    useEffect(() => {
        setValue("volumeOptions.subpathRequired", requiredPrefix, { shouldDirty: false, shouldValidate: false });
        if (requiredPrefix && !subpathValue) {
            setValue("volumeOptions.subpath", requiredPrefix, { shouldDirty: false, shouldValidate: false });
        }
    }, [requiredPrefix, setValue, subpathValue]);

    return (
        <>
            <Field>
                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="Volume"
                            isRequired
                        />
                    }
                >
                    <Select
                        {...volumeField}
                        value={volumeField.value}
                        onValueChange={volumeField.onChange}
                        disabled={isLoadingVolumes}
                    >
                        <SelectTrigger aria-invalid={volumeInvalid}>
                            <SelectValue placeholder={isLoadingVolumes ? "Loading volumes…" : "Volume"} />
                        </SelectTrigger>
                        <SelectContent>
                            {volumes.map(vol => (
                                <SelectItem
                                    key={vol.id}
                                    value={vol.id}
                                >
                                    {vol.name}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>

                    <FieldError errors={[volumeError]} />
                </InfoBlock>
            </Field>

            {requiredPrefix !== "" && (
                <Field>
                    <InfoBlock title={<LabelWithInfo label="Required Subpath Prefix" />}>
                        <div className={cn(dashedBorderBox, "py-2 px-3 rounded-md")}>{requiredPrefix}</div>
                    </InfoBlock>
                </Field>
            )}

            <Field>
                <InfoBlock title={<LabelWithInfo label="Subpath" />}>
                    <Input
                        {...subpathField}
                        id="subpath"
                        value={subpathValue}
                        placeholder={requiredPrefix}
                    />
                </InfoBlock>
            </Field>

            <Field>
                <InfoBlock title={<LabelWithInfo label="No Copy" />}>
                    <Checkbox
                        checked={noCopyField.value ?? false}
                        onCheckedChange={noCopyField.onChange}
                    />
                </InfoBlock>
            </Field>

            <Field>
                <InfoBlock title={<LabelWithInfo label="Labels" />}>
                    <KeyValueList<StorageMountFormInput>
                        name="volumeOptions.labels"
                        keyLabel="Key"
                        valueLabel="Value"
                    />
                </InfoBlock>
            </Field>
        </>
    );
}
