import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import { type AppConfigContainerSettingsFormSchemaInput } from "../schemas";

export function LabelsFields() {
    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Service Labels"
                        content="Key-value labels applied to the Docker service."
                    />
                }
            >
                <KeyValueList<AppConfigContainerSettingsFormSchemaInput>
                    name="serviceLabels"
                    className="max-w-[600px]"
                    checkDuplicates
                />
            </InfoBlock>
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Container Labels"
                        content="Key-value labels applied to the container spec."
                    />
                }
            >
                <KeyValueList<AppConfigContainerSettingsFormSchemaInput>
                    name="containerLabels"
                    className="max-w-[600px]"
                    checkDuplicates
                />
            </InfoBlock>
        </div>
    );
}
