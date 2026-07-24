import { Input } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import { type AppConfigResourcesFormSchemaInput, type AppConfigResourcesFormSchemaOutput } from "../schemas";

export function ResourceReservationFields() {
    const { control } = useFormContext<
        AppConfigResourcesFormSchemaInput,
        unknown,
        AppConfigResourcesFormSchemaOutput
    >();

    const { field: cpusField } = useController({ control, name: "reservations.cpus" });
    const { field: memoryField } = useController({ control, name: "reservations.memory" });

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="CPUs"
                        content="Number of CPUs reserved for the service."
                    />
                }
            >
                <InputNumber
                    value={cpusField.value}
                    onValueChange={val => {
                        cpusField.onChange(val);
                    }}
                    className="max-w-[100px]"
                    stepper={0.25}
                    min={0}
                    decimalScale={2}
                    fixedDecimalScale={false}
                />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Memory"
                        content="Amount of memory reserved for the service. Use DataSize values like 512mb or 1gb."
                    />
                }
            >
                <Input
                    value={memoryField.value ?? ""}
                    onChange={memoryField.onChange}
                    className="max-w-[100px]"
                    placeholder="512mb"
                />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Generic Resources"
                        content="User-defined resources such as GPUs or other accelerators."
                    />
                }
            >
                <KeyValueList<AppConfigResourcesFormSchemaInput>
                    name="reservations.genericResources"
                    keyField="kind"
                    keyLabel="Name"
                    keyPlaceholder="SSD"
                    valuePlaceholder="sda1 (string or integer)"
                    enableValueEditing
                    className="max-w-[590px]"
                />
            </InfoBlock>
        </div>
    );
}
