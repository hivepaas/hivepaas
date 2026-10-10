import { Input } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { KeyValueList } from "@application/shared/form";

import { type AppConfigResourcesFormSchemaInput, type AppConfigResourcesFormSchemaOutput } from "../schemas";

import { UlimitsFields } from "./ulimits-fields.com";

export function CapabilitiesFields() {
    const { control } = useFormContext<
        AppConfigResourcesFormSchemaInput,
        unknown,
        AppConfigResourcesFormSchemaOutput
    >();

    const { field: capabilityAddField } = useController({ control, name: "capabilities.capabilityAdd" });
    const { field: capabilityDropField } = useController({ control, name: "capabilities.capabilityDrop" });
    const { field: oomScoreAdjField } = useController({ control, name: "capabilities.oomScoreAdj" });

    return (
        <div className="flex flex-col gap-6">
            <div className="flex flex-col gap-2 rounded-md border border-amber-500/40 bg-amber-500/10 px-3 py-2 text-sm text-amber-700 dark:text-amber-400">
                <p>
                    <span className="font-medium">
                        Modifying container capabilities can introduce severe security risks.
                    </span>{" "}
                    Please make sure you understand the implications before proceeding.
                </p>
                <p>
                    Additionally, you must have <span className="font-medium">Write</span> permission on the{" "}
                    <span className="font-medium">Cluster</span> module to apply these changes.
                </p>
            </div>

            <UlimitsFields />

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Capabilities Add"
                        content="Linux capabilities to add to the container."
                    />
                }
            >
                <Input
                    value={capabilityAddField.value}
                    onChange={e => {
                        capabilityAddField.onChange(e.target.value);
                    }}
                    placeholder="SYS_ADMIN ANOTHER"
                    className="max-w-[500px]"
                />
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Capabilities Drop"
                        content="Linux capabilities to drop from the container."
                    />
                }
            >
                <Input
                    value={capabilityDropField.value}
                    onChange={e => {
                        capabilityDropField.onChange(e.target.value);
                    }}
                    placeholder="AUDIT_WRITE ANOTHER"
                    className="max-w-[500px]"
                />
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Out-of-Mem Score Adjustment"
                        content="Adjusts the OOM killer score for the container process."
                    />
                }
            >
                <InputNumber
                    value={oomScoreAdjField.value}
                    onValueChange={val => {
                        oomScoreAdjField.onChange(val);
                    }}
                    className="max-w-[100px]"
                />
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Sysctls"
                        content="Kernel parameters to set in the container namespace."
                    />
                }
            >
                <KeyValueList<AppConfigResourcesFormSchemaInput>
                    name="capabilities.sysctls"
                    keyField="name"
                    keyLabel="Name"
                    keyPlaceholder="net.core.somaxconn"
                    valuePlaceholder="1024"
                    enableValueEditing
                    className="max-w-[800px]"
                />
            </InfoBlock>
        </div>
    );
}
