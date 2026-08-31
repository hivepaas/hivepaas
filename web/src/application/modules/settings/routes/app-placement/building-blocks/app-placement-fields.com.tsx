import { Checkbox } from "@components/ui";
import { useController, useFormContext } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import type { SettingsAppPlacementFormSchemaInput, SettingsAppPlacementFormSchemaOutput } from "../schemas";

export function AppPlacementFields() {
    const { control } = useFormContext<
        SettingsAppPlacementFormSchemaInput,
        unknown,
        SettingsAppPlacementFormSchemaOutput
    >();

    const { field: excludeManagerNodes } = useController({ control, name: "excludeManagerNodes" });
    const { field: excludeBuildNodes } = useController({ control, name: "excludeBuildNodes" });

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Exclude Manager Nodes"
                        content="Avoid placing application workloads on Docker manager nodes when possible."
                    />
                }
            >
                <Checkbox
                    checked={excludeManagerNodes.value}
                    onCheckedChange={checked => {
                        excludeManagerNodes.onChange(checked === true);
                    }}
                />
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Exclude Build Nodes"
                        content="Avoid placing application workloads on nodes dedicated to image builds when possible."
                    />
                }
            >
                <Checkbox
                    checked={excludeBuildNodes.value}
                    onCheckedChange={checked => {
                        excludeBuildNodes.onChange(checked === true);
                    }}
                />
            </InfoBlock>
        </div>
    );
}
