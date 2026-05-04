import React, { useEffect, useMemo } from "react";

import { cn } from "@/lib/utils";
import { Input } from "@components/ui";
import { Field, FieldError } from "@components/ui/field";
import { SelectItem } from "@components/ui/select";
import { dashedBorderBox } from "@lib/styles";
import { useController, useFormContext, useWatch } from "react-hook-form";
import type { ProjectStorageSettings } from "~/projects/domain";

import { InfoBlock, InputWithAddOn, LabelWithInfo, SelectWithAddon } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form/key-value-list/key-value-list.com";

import type { StorageMountFormInput, StorageMountFormOutput } from "../../schemas";
import { computeRequiredSubpath } from "../../utils/required-subpath.util";

interface ClusterFieldsProps {
    projectRules?: ProjectStorageSettings;
    projectKey?: string;
    appLocalKey?: string;
}

export function ClusterFields({ projectRules, projectKey, appLocalKey }: ClusterFieldsProps) {
    const { control, setValue } = useFormContext<StorageMountFormInput, unknown, StorageMountFormOutput>();

    const {
        field: volumeField,
        fieldState: { invalid: volumeInvalid, error: volumeError },
    } = useController({ name: "clusterOptions.volume", control });
    const { field: subpathField } = useController({ name: "clusterOptions.subpath", control });
    const { field: noCopyField } = useController({ name: "clusterOptions.noCopy", control });

    const volumes = projectRules?.clusterVolumeSettings?.volumes ?? [];

    const requiredPrefix = useMemo(
        () =>
            computeRequiredSubpath(
                {
                    enabled: projectRules?.clusterVolumeSettings?.enabled,
                    baseSubpath: projectRules?.clusterVolumeSettings?.baseSubpath,
                    appsMustUseSubPaths: projectRules?.clusterVolumeSettings?.appsMustUseSubPaths,
                    hasItems: (projectRules?.clusterVolumeSettings?.volumes?.length ?? 0) > 0,
                },
                projectKey,
                appLocalKey,
            ),
        [projectRules?.clusterVolumeSettings, projectKey, appLocalKey],
    );

    const subpathValue = useWatch({ control, name: "clusterOptions.subpath" }) ?? "";
    useEffect(() => {
        setValue("clusterOptions.subpathRequired", requiredPrefix, { shouldDirty: false, shouldValidate: false });
        if (requiredPrefix && !subpathValue) {
            setValue("clusterOptions.subpath", requiredPrefix, { shouldDirty: false, shouldValidate: false });
        }
    }, [requiredPrefix, setValue, subpathValue]);

    return (
        <>
            <Field>
                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="Cluster Volume"
                            isRequired
                        />
                    }
                >
                    {volumes.length > 0 ? (
                        <SelectWithAddon
                            {...volumeField}
                            addonLeft="Volume"
                            value={volumeField.value || ""}
                            onValueChange={volumeField.onChange}
                        >
                            {volumes.map((vol: { id: string; name: string }) => (
                                <SelectItem
                                    key={vol.id}
                                    value={vol.id}
                                >
                                    {vol.name}
                                </SelectItem>
                            ))}
                        </SelectWithAddon>
                    ) : (
                        <InputWithAddOn
                            {...volumeField}
                            id="volume"
                            placeholder="cluster-volume-name"
                            aria-invalid={volumeInvalid}
                            addonLeft="Volume"
                        />
                    )}
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
                    <label className="flex items-center gap-2">
                        <input
                            type="checkbox"
                            checked={noCopyField.value ?? false}
                            onChange={noCopyField.onChange}
                        />
                        <span className="text-sm">No Copy</span>
                    </label>
                </InfoBlock>
            </Field>

            <Field>
                <InfoBlock title={<LabelWithInfo label="Labels" />}>
                    <KeyValueList
                        name="clusterOptions.labels"
                        keyLabel="Key"
                        valueLabel="Value"
                    />
                </InfoBlock>
            </Field>
        </>
    );
}
