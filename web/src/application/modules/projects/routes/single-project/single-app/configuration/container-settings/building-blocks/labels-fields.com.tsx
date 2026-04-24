import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import { type AppConfigContainerSettingsFormSchemaInput } from "../schemas";

export function LabelsFields() {
    return (
        <>
            <h3 className="font-medium bg-accent py-2 px-3 rounded-lg">Labels</h3>
            <div className="flex flex-col gap-6 px-2">
                <InfoBlock
                    title={
                        <LabelWithInfo
                            label="Labels"
                            content="Key-value labels applied to the container."
                        />
                    }
                >
                    <KeyValueList<AppConfigContainerSettingsFormSchemaInput>
                        name="labels"
                        className="max-w-[600px]"
                        checkDuplicates
                    />
                </InfoBlock>
            </div>
        </>
    );
}
