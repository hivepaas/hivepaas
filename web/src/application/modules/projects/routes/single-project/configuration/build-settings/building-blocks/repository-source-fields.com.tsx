import { Checkbox } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { ProjectImageBuildSettingsFormSchemaInput, ProjectImageBuildSettingsFormSchemaOutput } from "../schemas";

export function RepositorySourceFields({ cacheNote }: RepositorySourceFieldsProps) {
    const { control } = useFormContext<
        ProjectImageBuildSettingsFormSchemaInput,
        unknown,
        ProjectImageBuildSettingsFormSchemaOutput
    >();

    const { field: repoCacheField } = useController({ control, name: "sources.repoCache" });

    return (
        <div className="flex flex-col gap-6">
            {cacheNote}

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Source Cache"
                        content="Cache repository sources between image builds."
                    />
                }
            >
                <Checkbox
                    checked={repoCacheField.value}
                    onCheckedChange={checked => {
                        repoCacheField.onChange(checked === true);
                    }}
                />
            </InfoBlock>
        </div>
    );
}

interface RepositorySourceFieldsProps {
    cacheNote?: React.ReactNode;
}
