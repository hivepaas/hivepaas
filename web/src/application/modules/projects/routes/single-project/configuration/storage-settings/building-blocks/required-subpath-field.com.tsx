import { Input } from "@components/ui";
import { type FieldPath, useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { ProjectStorageSettingsFormSchemaInput, ProjectStorageSettingsFormSchemaOutput } from "../schemas";

import {
    STORAGE_SETTINGS_CONTROL_CLASS_NAME,
    STORAGE_SETTINGS_FIELD_TITLE_WIDTH,
} from "./storage-settings-layout.constants";

type SubpathFieldPath = Extract<
    FieldPath<ProjectStorageSettingsFormSchemaInput>,
    "bindSettings.subpathTemplate" | "volumeSettings.subpathTemplate" | "clusterVolumeSettings.subpathTemplate"
>;

export function RequiredSubpathField({ name }: Props) {
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
                    label="Required Subpath"
                    content="Subpath template that apps must use for this storage type."
                />
            }
        >
            <div className={STORAGE_SETTINGS_CONTROL_CLASS_NAME}>
                <Input
                    value={field.value}
                    onChange={field.onChange}
                    className="max-w-[460px]"
                    placeholder="project_data/{{project}}/{{env}}/{{app}}"
                />
            </div>
        </InfoBlock>
    );
}

type Props = {
    name: SubpathFieldPath;
};
