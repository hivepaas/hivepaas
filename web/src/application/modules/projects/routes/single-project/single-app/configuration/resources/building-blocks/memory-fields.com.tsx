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
    const { field: shmSizeField } = useController({ control, name: "memory.shmSize" });

    return (
        <div className="flex flex-col gap-6">
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

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Shm Size"
                        content="Size of the tmpfs mounted to /dev/shm. Use DataSize values like 1gb."
                    />
                }
            >
                <div className="flex items-center gap-6">
                    <Input
                        value={shmSizeField.value ?? ""}
                        onChange={shmSizeField.onChange}
                        className="max-w-[100px]"
                        placeholder="1gb"
                    />
                    <span className="text-sm">This will mount a tmpfs to /dev/shm.</span>
                </div>
            </InfoBlock>
        </div>
    );
}
