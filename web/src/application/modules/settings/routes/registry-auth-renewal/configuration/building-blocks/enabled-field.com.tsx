import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { Checkbox } from "@/components/ui";

import type {
    SystemRegistryAuthRenewalConfigurationFormInput,
    SystemRegistryAuthRenewalConfigurationFormOutput,
} from "../schemas";

type SchemaInput = SystemRegistryAuthRenewalConfigurationFormInput;
type SchemaOutput = SystemRegistryAuthRenewalConfigurationFormOutput;

export function EnabledField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field: status } = useController({ control, name: "status" });

    return (
        <InfoBlock
            titleWidth={220}
            title="Enabled"
        >
            <Checkbox
                checked={status.value === ESettingStatus.Active}
                onCheckedChange={checked => {
                    status.onChange(checked ? ESettingStatus.Active : ESettingStatus.Disabled);
                }}
            />
        </InfoBlock>
    );
}
