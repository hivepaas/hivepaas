import React from "react";

import { cn } from "@/lib/utils";
import { Checkbox, Input } from "@components/ui";
import { Field, FieldError } from "@components/ui/field";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { dashedBorderBox } from "@lib/styles";
import { useController, useFormContext, useWatch } from "react-hook-form";
import { useParams } from "react-router";
import { ProjectDockerVolumesQueries } from "~/projects/data/queries/project-docker-volumes";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import type { SettingStorageSettings } from "~/settings/domain";

import { AppLink, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import type { StorageMountFormInput, StorageMountFormOutput } from "../../schemas";

interface VolumeFieldsProps {
    storageSettings?: SettingStorageSettings;
    projectKey?: string;
    appLocalKey?: string;
}

type MountVariant = "volume" | "cluster";
type MountOptionsPath = "volumeOptions" | "clusterOptions";
type MountVolumePath = `${MountOptionsPath}.volume`;
type MountSubpathPath = `${MountOptionsPath}.subpath`;
type MountNoCopyPath = `${MountOptionsPath}.noCopy`;
type MountSubpathRequiredPath = `${MountOptionsPath}.subpathRequired`;

function VolumeOrClusterMountFields({ variant }: VolumeFieldsProps & { variant: MountVariant }) {
    const { control } = useFormContext<StorageMountFormInput, unknown, StorageMountFormOutput>();
    const { id: projectId } = useParams<{ id: string }>();
    const optionsPath: MountOptionsPath = variant === "cluster" ? "clusterOptions" : "volumeOptions";
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
    const { field: subpathRequiredField } = useController({ name: subpathRequiredPath, control });

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

    const subpathValue = useWatch({ control, name: subpathPath }) ?? "";

    return (
        <>
            <InfoBlock
                title={
                    <LabelWithInfo
                        label={volumeLabel}
                        isRequired
                    />
                }
                titleWidth={180}
            >
                <Field>
                    <Select
                        {...volumeField}
                        value={volumeField.value}
                        onValueChange={volumeField.onChange}
                        disabled={isLoadingVolumes}
                    >
                        <SelectTrigger
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
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

                    <div className="text-xs">
                        <AppLink.Basic
                            to={ROUTE.projects.single.clusterResources.volumes.$route(projectId ?? "")}
                            className="text-link"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            Configure Volumes
                        </AppLink.Basic>
                    </div>
                </Field>
            </InfoBlock>

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
