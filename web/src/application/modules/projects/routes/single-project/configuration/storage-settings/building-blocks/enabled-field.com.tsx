import { Checkbox } from "@components/ui";
import { type FieldPath, useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { ProjectStorageSettingsFormSchemaInput, ProjectStorageSettingsFormSchemaOutput } from "../schemas";

import { STORAGE_SETTINGS_FIELD_TITLE_WIDTH } from "./storage-settings-layout.constants";

type BooleanFieldPath = Extract<
    FieldPath<ProjectStorageSettingsFormSchemaInput>,
    "bindSettings.enabled" | "volumeSettings.enabled" | "clusterVolumeSettings.enabled" | "tmpfsSettings.enabled"
>;

export function EnabledField({ name }: Props) {
    const { control } = useFormContext<
        ProjectStorageSettingsFormSchemaInput,
        unknown,
        ProjectStorageSettingsFormSchemaOutput
    >();
    const { field } = useController({ control, name });

    return (
        <InfoBlock
            titleWidth={STORAGE_SETTINGS_FIELD_TITLE_WIDTH}
            title={
                <LabelWithInfo
                    label="Enabled"
                    content="Enable this storage constraint for project apps."
                />
            }
        >
            <Checkbox
                checked={field.value}
                onCheckedChange={field.onChange}
            />
        </InfoBlock>
    );
}

type Props = {
    name: BooleanFieldPath;
};
