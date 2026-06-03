import { Input } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { ProjectStorageSettingsFormSchemaInput, ProjectStorageSettingsFormSchemaOutput } from "../schemas";

import {
    STORAGE_SETTINGS_CONTROL_CLASS_NAME,
    STORAGE_SETTINGS_FIELD_TITLE_WIDTH,
} from "./storage-settings-layout.constants";

export function TmpfsMaxSizeField() {
    const { control } = useFormContext<
        ProjectStorageSettingsFormSchemaInput,
        unknown,
        ProjectStorageSettingsFormSchemaOutput
    >();
    const { field } = useController({ control, name: "tmpfsSettings.maxSize" });

    return (
        <InfoBlock
            titleWidth={STORAGE_SETTINGS_FIELD_TITLE_WIDTH}
            title={
                <LabelWithInfo
                    label="Max Size"
                    content="Maximum tmpfs size allowed for apps in this project. Use DataSize values like 2gb."
                />
            }
        >
            <div className={STORAGE_SETTINGS_CONTROL_CLASS_NAME}>
                <Input
                    value={field.value}
                    onChange={field.onChange}
                    className="max-w-[460px]"
                    placeholder="2gb"
                />
            </div>
        </InfoBlock>
    );
}
