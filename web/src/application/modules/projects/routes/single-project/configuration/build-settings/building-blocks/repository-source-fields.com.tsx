import { Checkbox } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { ProjectImageBuildSettingsFormSchemaInput, ProjectImageBuildSettingsFormSchemaOutput } from "../schemas";

export function RepositorySourceFields({ cacheNote }: RepositorySourceFieldsProps) {
    const { control } = useFormContext<
        ProjectImageBuildSettingsFormSchemaInput,
        unknown,
        ProjectImageBuildSettingsFormSchemaOutput
    >();

    const { field: checkoutMaxDepthField } = useController({ control, name: "sources.checkoutMaxDepth" });
    const { field: repoCacheField } = useController({ control, name: "sources.repoCache" });

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Checkout Max Depth"
                        content="Maximum Git checkout depth for repository sources. Use 0 or leave empty for full depth."
                    />
                }
            >
                <InputNumber
                    value={checkoutMaxDepthField.value}
                    onValueChange={checkoutMaxDepthField.onChange}
                    className="max-w-[100px]"
                    min={0}
                    decimalScale={0}
                    fixedDecimalScale={false}
                />
            </InfoBlock>

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
