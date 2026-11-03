import { type ReactNode } from "react";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import { type AppConfigContainerSettingsFormSchemaInput } from "../schemas";

/** The service's and container's labels. toolbar sits at the top of the section: the reveal of system labels. */
export function LabelsFields({ toolbar }: { toolbar?: ReactNode }) {
    return (
        <div className="flex flex-col gap-6">
            {toolbar && <div className="flex items-center gap-2">{toolbar}</div>}
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Service Labels"
                        content="Key-value labels applied to the Docker service."
                    />
                }
            >
                <KeyValueList<AppConfigContainerSettingsFormSchemaInput>
                    name="serviceLabels"
                    className="max-w-[800px]"
                    checkDuplicates
                    enableValueEditing
                    ratio="55-45"
                />
            </InfoBlock>
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Container Labels"
                        content="Key-value labels applied to the container spec."
                    />
                }
            >
                <KeyValueList<AppConfigContainerSettingsFormSchemaInput>
                    name="containerLabels"
                    className="max-w-[800px]"
                    checkDuplicates
                    enableValueEditing
                    ratio="55-45"
                />
            </InfoBlock>
        </div>
    );
}
