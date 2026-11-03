import { Input } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { DOCS_URL } from "@application/shared/constants";
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
                titleWidth={220}
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
                titleWidth={220}
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
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Generic Resources"
                        content="Resources a node advertises, by name and count - such as NVIDIA-GPU, 2 for two GPUs - or one by its name."
                    />
                }
            >
                <KeyValueList<AppConfigResourcesFormSchemaInput>
                    name="reservations.genericResources"
                    keyField="kind"
                    keyLabel="Name"
                    keyPlaceholder="NVIDIA-GPU"
                    valuePlaceholder="2 (a count, or a name)"
                    enableValueEditing
                    className="max-w-[800px]"
                />
                <div className="mt-3 flex max-w-[800px] flex-col gap-1 rounded-md border border-sky-500/30 bg-sky-500/10 px-3 py-2 text-sm text-sky-700 dark:text-sky-400">
                    <p>
                        <span className="font-medium">To give the app a GPU</span>, add the name its node lists its GPUs
                        as, with how many: <code>NVIDIA-GPU</code> or <code>AMD_GPU</code>, and <code>1</code>. The app
                        then runs only on a node with that many free.
                    </p>
                    <p>
                        Each such node needs its GPU maker&apos;s container runtime:{" "}
                        <a
                            href={`${DOCS_URL}/configuring-apps/resources-and-placement#gpus`}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="underline underline-offset-2"
                        >
                            how to prepare a node
                        </a>
                    </p>
                    <p>
                        Reserving a GPU needs <span className="font-medium">Write</span> permission on the{" "}
                        <span className="font-medium">Cluster</span> module.
                    </p>
                </div>
            </InfoBlock>
        </div>
    );
}
