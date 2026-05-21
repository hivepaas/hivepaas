import { Input } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type AppConfigResourcesFormSchemaInput, type AppConfigResourcesFormSchemaOutput } from "../schemas";

export function MemoryFields() {
    const { control } = useFormContext<
        AppConfigResourcesFormSchemaInput,
        unknown,
        AppConfigResourcesFormSchemaOutput
    >();

    const { field: swapField } = useController({ control, name: "memory.swap" });
    const { field: swappinessField } = useController({ control, name: "memory.swappiness" });

    return (
        <div className="flex flex-col gap-6 px-2">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Swap Memory"
                        content="Total memory plus swap available to the service. Use DataSize values like 10gb."
                    />
                }
            >
                <Input
                    value={swapField.value ?? ""}
                    onChange={swapField.onChange}
                    className="max-w-[100px]"
                    placeholder="-1, 10gb"
                />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Swappiness"
                        content="Controls how aggressively the kernel swaps container memory pages."
                    />
                }
            >
                <InputNumber
                    value={swappinessField.value}
                    onValueChange={val => {
                        swappinessField.onChange(val);
                    }}
                    className="max-w-[100px]"
                    decimalScale={0}
                    fixedDecimalScale={false}
                />
            </InfoBlock>
        </div>
    );
}
