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

import type { StorageMountFormInput, StorageMountFormOutput } from "../../schemas";
import { computeRequiredSubpath } from "../../utils/required-subpath.util";

interface VolumeFieldsProps {
    projectRules?: ProjectStorageSettings;
    projectKey?: string;
    appLocalKey?: string;
}

type MountVariant = "volume" | "cluster";
type MountOptionsPath = "volumeOptions" | "clusterOptions";
type MountVolumePath = `${MountOptionsPath}.volume`;
type MountSubpathPath = `${MountOptionsPath}.subpath`;
type MountNoCopyPath = `${MountOptionsPath}.noCopy`;
type MountSubpathRequiredPath = `${MountOptionsPath}.subpathRequired`;

function VolumeOrClusterMountFields({
    projectRules,
    projectKey,
    appLocalKey,
    variant,
}: VolumeFieldsProps & { variant: MountVariant }) {
    const { control, setValue } = useFormContext<StorageMountFormInput, unknown, StorageMountFormOutput>();
    const { id: projectId } = useParams<{ id: string }>();
    const optionsPath: MountOptionsPath = variant === "cluster" ? "clusterOptions" : "volumeOptions";
    const rulesSettings = variant === "cluster" ? projectRules?.clusterVolumeSettings : projectRules?.volumeSettings;
    const volumeLabel = variant === "cluster" ? "Cluster Volume" : "Volume";

    const volumePath: MountVolumePath = `${optionsPath}.volume`;
    const subpathPath: MountSubpathPath = `${optionsPath}.subpath`;
    const noCopyPath: MountNoCopyPath = `${optionsPath}.noCopy`;
    const subpathRequiredPath: MountSubpathRequiredPath = `${optionsPath}.subpathRequired`;

    const {
        field: volumeField,
        fieldState: { invalid: volumeInvalid, error: volumeError },
    } = useController({ name: volumePath, control });
    const { field: subpathField } = useController({ name: subpathPath, control });
    const { field: noCopyField } = useController({ name: noCopyPath, control });

    const { data: volumesData, isLoading: isLoadingVolumes } = ProjectDockerVolumesQueries.useList(
        {
            projectID: projectId ?? "",
            type: variant,
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
                    enabled: rulesSettings?.enabled,
                    baseSubpath: rulesSettings?.baseSubpath,
                    appsMustUseSubPaths: rulesSettings?.appsMustUseSubPaths,
                },
                projectKey,
                appLocalKey,
            ),
        [rulesSettings, projectKey, appLocalKey],
    );

    const subpathValue = useWatch({ control, name: subpathPath }) ?? "";
    useEffect(() => {
        setValue(subpathRequiredPath, requiredPrefix, { shouldDirty: false, shouldValidate: false });
        if (requiredPrefix && !subpathValue) {
            setValue(subpathPath, requiredPrefix, { shouldDirty: false, shouldValidate: false });
        }
    }, [requiredPrefix, setValue, subpathRequiredPath, subpathPath, subpathValue]);

    return (
        <>
            <Field>
                <InfoBlock
                    title={
                        <LabelWithInfo
                            label={volumeLabel}
                            isRequired
                        />
                    }
                    titleWidth={180}
                >
                    <Select
                        {...volumeField}
                        value={volumeField.value}
                        onValueChange={volumeField.onChange}
                        disabled={isLoadingVolumes}
                    >
                        <SelectTrigger
                            className="w-[220px]"
                            aria-invalid={volumeInvalid}
                        >
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
                    title={<LabelWithInfo label="No Copy" />}
                    titleWidth={180}
                >
                    <Checkbox
                        checked={noCopyField.value ?? false}
                        onCheckedChange={noCopyField.onChange}
                    />
                </InfoBlock>
            </Field>
        </>
    );
}

export function VolumeFields(props: VolumeFieldsProps) {
    return (
        <VolumeOrClusterMountFields
            {...props}
            variant="volume"
        />
    );
}

export function ClusterFields(props: VolumeFieldsProps) {
    return (
        <VolumeOrClusterMountFields
            {...props}
            variant="cluster"
        />
    );
}
