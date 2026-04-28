import React from "react";

import { Field, FieldError, FieldLabel } from "@components/ui/field";
import { SelectItem } from "@components/ui/select";
import { useController, useFormContext } from "react-hook-form";
import type { ProjectStorageSettings } from "~/projects/domain";

import { InputWithAddOn, SelectWithAddon } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form/key-value-list/key-value-list.com";

import type { StorageMountFormInput, StorageMountFormOutput } from "../../schemas";

interface ClusterFieldsProps {
    projectRules?: ProjectStorageSettings;
}

export function ClusterFields({ projectRules }: ClusterFieldsProps) {
    const { control } = useFormContext<StorageMountFormInput, unknown, StorageMountFormOutput>();

    const {
        field: volumeField,
        fieldState: { invalid: volumeInvalid, error: volumeError },
    } = useController({ name: "clusterOptions.volume", control });
    const { field: subpathField } = useController({ name: "clusterOptions.subpath", control });
    const { field: noCopyField } = useController({ name: "clusterOptions.noCopy", control });

    const volumes = projectRules?.clusterVolumeSettings?.volumes ?? [];

    return (
        <>
            <Field>
                <FieldLabel htmlFor="volume">Cluster Volume *</FieldLabel>
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
            </Field>

            <Field>
                <FieldLabel htmlFor="subpath">Subpath</FieldLabel>
                <InputWithAddOn
                    {...subpathField}
                    id="subpath"
                    placeholder="app/data"
                    addonLeft="Subpath"
                />
            </Field>

            <Field>
                <label className="flex items-center gap-2">
                    <input
                        type="checkbox"
                        checked={noCopyField.value ?? false}
                        onChange={noCopyField.onChange}
                    />
                    <span className="text-sm">No Copy</span>
                </label>
            </Field>

            <Field>
                <FieldLabel>Labels</FieldLabel>
                <KeyValueList
                    name="clusterOptions.labels"
                    keyLabel="Key"
                    valueLabel="Value"
                />
            </Field>
        </>
    );
}
