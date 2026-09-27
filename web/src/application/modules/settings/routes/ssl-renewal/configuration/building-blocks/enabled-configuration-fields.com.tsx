import { useFormContext, useWatch } from "react-hook-form";

import { ESettingStatus } from "@application/shared/enums";

import type { SystemSslRenewalConfigurationFormInput, SystemSslRenewalConfigurationFormOutput } from "../schemas";

import { GeneralFields } from "./general-fields.com";
import { NotificationFields } from "./notification-fields.com";

type SchemaInput = SystemSslRenewalConfigurationFormInput;
type SchemaOutput = SystemSslRenewalConfigurationFormOutput;

export function EnabledConfigurationFields({ nextRuns, readOnly }: Props) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const status = useWatch({ control, name: "status" });

    if (status !== ESettingStatus.Active) {
        return null;
    }

    return (
        <>
            <GeneralFields nextRuns={nextRuns} />
            <NotificationFields readOnly={readOnly} />
        </>
    );
}

interface Props {
    nextRuns: Date[];
    readOnly: boolean;
}
