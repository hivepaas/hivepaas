import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { AlertTriangle } from "lucide-react";
import { useFormContext, useWatch } from "react-hook-form";

import { ESettingStatus } from "@application/shared/enums";

import type {
    SystemRegistryAuthRenewalConfigurationFormInput,
    SystemRegistryAuthRenewalConfigurationFormOutput,
} from "../schemas";

import { GeneralFields } from "./general-fields.com";
import { NotificationFields } from "./notification-fields.com";

type SchemaInput = SystemRegistryAuthRenewalConfigurationFormInput;
type SchemaOutput = SystemRegistryAuthRenewalConfigurationFormOutput;

export function EnabledConfigurationFields({ nextRuns, readOnly }: Props) {
    const { control } = useFormContext<SchemaInput, unknown, SchemaOutput>();
    const status = useWatch({ control, name: "status" });

    if (status !== ESettingStatus.Active) {
        return (
            <div className={cn(dashedBorderBox, "flex items-start gap-2")}>
                <AlertTriangle className="mt-0.5 size-4 shrink-0 text-amber-600 dark:text-amber-500" />
                <span>
                    With the renewal off, the apps that pull their image with an Amazon ECR credential cannot be started
                    on another node, or after their image is removed from theirs, once 12 hours have passed since their
                    last deployment.
                </span>
            </div>
        );
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
