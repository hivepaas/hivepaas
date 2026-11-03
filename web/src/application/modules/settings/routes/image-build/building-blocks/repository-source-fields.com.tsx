import type { ReactNode } from "react";

import { Checkbox } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { SettingsImageBuildFormSchemaInput, SettingsImageBuildFormSchemaOutput } from "../schemas";

export function RepositorySourceFields({ cacheNote }: RepositorySourceFieldsProps) {
    const { control } = useFormContext<
        SettingsImageBuildFormSchemaInput,
        unknown,
        SettingsImageBuildFormSchemaOutput
    >();

    const { field: repoCacheField } = useController({ control, name: "sources.repoCache" });

    return (
        <div className="flex flex-col gap-6">
            {cacheNote}

            <InfoBlock
                titleWidth={220}
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
    cacheNote?: ReactNode;
}
