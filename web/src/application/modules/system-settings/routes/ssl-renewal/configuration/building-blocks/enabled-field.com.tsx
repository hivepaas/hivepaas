import { useController, useFormContext } from "react-hook-form";

import { InfoBlock } from "@application/shared/components";
import { ESettingStatus } from "@application/shared/enums";

import { Checkbox } from "@/components/ui";

import type { SystemSslRenewalConfigurationFormInput, SystemSslRenewalConfigurationFormOutput } from "../schemas";

type SchemaInput = SystemSslRenewalConfigurationFormInput;
type SchemaOutput = SystemSslRenewalConfigurationFormOutput;

export function EnabledField() {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const { field: status } = useController({ control, name: "status" });

    return (
        <InfoBlock title="Enabled">
            <Checkbox
                checked={status.value === ESettingStatus.Active}
                onCheckedChange={checked => {
                    status.onChange(checked ? ESettingStatus.Active : ESettingStatus.Disabled);
                }}
            />
        </InfoBlock>
    );
}
