import { Input } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { ProjectImageBuildSettingsFormSchemaInput, ProjectImageBuildSettingsFormSchemaOutput } from "../schemas";

export function ResourceLimitFields() {
    const { control } = useFormContext<
        ProjectImageBuildSettingsFormSchemaInput,
        unknown,
        ProjectImageBuildSettingsFormSchemaOutput
    >();

    const { field: cpusField } = useController({ control, name: "resources.cpus" });
    const { field: memField } = useController({ control, name: "resources.mem" });
    const { field: memSwapField } = useController({ control, name: "resources.memSwap" });
    const { field: shmSizeField } = useController({ control, name: "resources.shmSize" });

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                title={
                    <LabelWithInfo
                        label="CPUs"
                        content="Maximum CPUs used by the image build process."
                    />
                }
            >
                <InputNumber
                    value={cpusField.value}
                    onValueChange={cpusField.onChange}
                    className="max-w-[100px]"
                    min={0}
                    decimalScale={0}
                    fixedDecimalScale={false}
                />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Memory"
                        content="Maximum memory used by the image build process. Use DataSize values like 4gb."
                    />
                }
            >
                <Input
                    value={memField.value ?? ""}
                    onChange={memField.onChange}
                    className="max-w-[100px]"
                    placeholder="4gb"
                />
            </InfoBlock>

            <InfoBlock
                title={
                    <LabelWithInfo
                        label="Swap Memory"
                        content="Total memory plus swap available to the image build process. Use DataSize values like 8gb."
                    />
                }
            >
                <Input
                    value={memSwapField.value ?? ""}
                    onChange={memSwapField.onChange}
                    className="max-w-[100px]"
                    placeholder="8gb"
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
                <Input
                    value={shmSizeField.value ?? ""}
                    onChange={shmSizeField.onChange}
                    className="max-w-[100px]"
                    placeholder="1gb"
                />
            </InfoBlock>
        </div>
    );
}
