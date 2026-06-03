import { FieldError } from "@components/ui";
import { useFormState } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { SingleValueList } from "@application/shared/form";

import type { ProjectStorageSettingsFormSchemaInput } from "../schemas";

import {
    STORAGE_SETTINGS_CONTROL_CLASS_NAME,
    STORAGE_SETTINGS_FIELD_TITLE_WIDTH,
} from "./storage-settings-layout.constants";

export function AllowedBaseDirectoriesFields({ readOnly = false }: Props) {
    const { errors } = useFormState<ProjectStorageSettingsFormSchemaInput>();

    return (
        <InfoBlock
            titleWidth={STORAGE_SETTINGS_FIELD_TITLE_WIDTH}
            title={
                <LabelWithInfo
                    label="Allowed Base Directories"
                    content="Directories that apps in this project are allowed to bind mount."
                />
            }
        >
            <div className={`flex flex-col gap-2 ${STORAGE_SETTINGS_CONTROL_CLASS_NAME}`}>
                <SingleValueList<ProjectStorageSettingsFormSchemaInput>
                    name="bindSettings.baseDirs"
                    label="Directory"
                    placeholder="directory"
                    className="max-w-[545px]"
                    disabled={readOnly}
                />
                <FieldError errors={[errors.bindSettings?.baseDirs]} />
            </div>
        </InfoBlock>
    );
}

type Props = {
    readOnly?: boolean;
};
