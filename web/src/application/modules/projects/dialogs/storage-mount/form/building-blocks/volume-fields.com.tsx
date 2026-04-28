import React from "react";

import { Checkbox } from "@components/ui";
import { Field, FieldError } from "@components/ui/field";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@components/ui/select";
import { useController, useFormContext } from "react-hook-form";
import { useParams } from "react-router-dom";
import { ProjectDockerVolumesQueries } from "~/projects/data/queries/project-docker-volumes";
import type { ProjectStorageSettings } from "~/projects/domain";

import { InfoBlock, InputWithAddOn, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form/key-value-list/key-value-list.com";

import type { StorageMountFormInput, StorageMountFormOutput } from "../../schemas";

interface VolumeFieldsProps {
    projectRules?: ProjectStorageSettings;
}

export function VolumeFields({ projectRules: _projectRules }: VolumeFieldsProps) {
    const { control } = useFormContext<StorageMountFormInput, unknown, StorageMountFormOutput>();
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
                    >
                        <SelectTrigger>
                            <SelectValue placeholder="Volume" />
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

            <Field>
                <InfoBlock title={<LabelWithInfo label="Subpath" />}>
                    <InputWithAddOn
                        {...subpathField}
                        id="subpath"
                        placeholder="app/data"
                        addonLeft="Subpath"
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
                    <KeyValueList
                        name="volumeOptions.labels"
                        keyLabel="Key"
                        valueLabel="Value"
                    />
                </InfoBlock>
            </Field>
        </>
    );
}
