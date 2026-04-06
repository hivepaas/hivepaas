import { InputNumber } from "@components/ui/input-number";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type AppConfigResourcesFormSchemaInput, type AppConfigResourcesFormSchemaOutput } from "../schemas";

export function ResourceLimitFields() {
    const { control } = useFormContext<
        AppConfigResourcesFormSchemaInput,
        unknown,
        AppConfigResourcesFormSchemaOutput
    >();

    const { field: cpusField } = useController({ control, name: "limits.cpus" });
    const { field: memoryField } = useController({ control, name: "limits.memoryMB" });
    const { field: pidsField } = useController({ control, name: "limits.pids" });

    return (
        <div className="flex flex-col gap-6 px-2">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="CPUs"
                        content="Maximum number of CPUs the service can use."
                    />
                }
            >
                <InputNumber
                    value={cpusField.value}
                    onValueChange={val => {
                        cpusField.onChange(val);
                    }}
                    className="max-w-[100px]"
                    min={0}
                    decimalScale={2}
                    fixedDecimalScale={false}
                />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Memory MB"
                        content="Maximum memory (in MB) the service can use."
                    />
                }
            >
                <InputNumber
                    value={memoryField.value}
                    onValueChange={val => {
                        memoryField.onChange(val);
                    }}
                    className="max-w-[100px]"
                    min={0}
                />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Pids"
                        content="Maximum number of PIDs (processes) the service can create."
                    />
                }
            >
                <InputNumber
                    value={pidsField.value}
                    onValueChange={val => {
                        pidsField.onChange(val);
                    }}
                    className="max-w-[100px]"
                    min={0}
                />
            </InfoBlock>
        </div>
    );
}
