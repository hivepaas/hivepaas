import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import type { SystemRegistryAuthRenewal_UpdateOne_Req } from "~/system-settings/api/services";
import { SystemRegistryAuthRenewalCommands, SystemRegistryAuthRenewalQueries } from "~/system-settings/data";
import type { SystemRegistryAuthRenewalSettings } from "~/system-settings/domain";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { SystemRegistryAuthRenewalConfigurationForm } from "../form";
import type { SystemRegistryAuthRenewalConfigurationFormOutput } from "../schemas";
import type { SystemRegistryAuthRenewalConfigurationFormRef } from "../types";

type UpdatePayload = SystemRegistryAuthRenewal_UpdateOne_Req["data"]["payload"];

function mapFormValuesToPayload(
    values: SystemRegistryAuthRenewalConfigurationFormOutput,
    settings?: SystemRegistryAuthRenewalSettings,
): UpdatePayload {
    return {
        updateVer: settings?.updateVer ?? 0,
        status: values.status,
        schedule: {
            interval: values.scheduleInterval,
            ...(values.scheduleFrom ? { initialTime: values.scheduleFrom } : {}),
        },
        notification: {
            successUseDefault: values.notification.successUseDefault,
            success: {
                id: values.notification.successUseDefault ? "" : (values.notification.success?.id ?? ""),
            },
            failureUseDefault: values.notification.failureUseDefault,
            failure: {
                id: values.notification.failureUseDefault ? "" : (values.notification.failure?.id ?? ""),
            },
        },
    };
}

export function SettingsRegistryAuthRenewalConfigurationRoute() {
    const formRef = useRef<SystemRegistryAuthRenewalConfigurationFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const { data, isLoading } = SystemRegistryAuthRenewalQueries.useFindOne();

    const { mutate: update, isPending } = SystemRegistryAuthRenewalCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Registry auth renewal settings updated");
        },
    });

    function handleSubmit(values: SystemRegistryAuthRenewalConfigurationFormOutput) {
        if (!canWrite) {
            return;
        }

        update({
            payload: mapFormValuesToPayload(values, data?.data),
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <SystemRegistryAuthRenewalConfigurationForm
            ref={formRef}
            defaultValues={data?.data}
            onSubmit={handleSubmit}
            readOnly={!canWrite}
        >
            <FormActionBar>
                <PermissionTooltipAction
                    id={MODULE_IDS.System}
                    action="write"
                >
                    {({ isDenied }) => (
                        <Button
                            type="submit"
                            className="min-w-[100px]"
                            disabled={isPending || isDenied}
                            isLoading={isPending}
                        >
                            Save
                        </Button>
                    )}
                </PermissionTooltipAction>
            </FormActionBar>
        </SystemRegistryAuthRenewalConfigurationForm>
    );
}
