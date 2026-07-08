import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import type { SystemSslRenewal_UpdateOne_Req } from "~/system-settings/api/services";
import { SystemSslRenewalCommands, SystemSslRenewalQueries } from "~/system-settings/data";
import type { SystemSslRenewalSettings } from "~/system-settings/domain";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { SystemSslRenewalConfigurationForm } from "../form";
import { type SystemSslRenewalConfigurationFormOutput, SystemSslRenewalScheduleMode } from "../schemas";
import type { SystemSslRenewalConfigurationFormRef } from "../types";

type UpdatePayload = SystemSslRenewal_UpdateOne_Req["data"]["payload"];

function mapFormValuesToPayload(
    values: SystemSslRenewalConfigurationFormOutput,
    settings?: SystemSslRenewalSettings,
): UpdatePayload {
    return {
        updateVer: settings?.updateVer ?? 0,
        status: values.status,
        schedule: {
            interval: values.scheduleMode === SystemSslRenewalScheduleMode.Interval ? values.scheduleInterval : "",
            cronExpr: values.scheduleMode === SystemSslRenewalScheduleMode.Cron ? values.scheduleCronExpr : "",
            initialTime: values.scheduleFrom ?? null,
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

export function SystemSettingsSslRenewalConfigurationRoute() {
    const formRef = useRef<SystemSslRenewalConfigurationFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const { data, isLoading } = SystemSslRenewalQueries.useFindOne();

    const { mutate: update, isPending } = SystemSslRenewalCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("SSL renewal settings updated");
        },
    });

    function handleSubmit(values: SystemSslRenewalConfigurationFormOutput) {
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
        <SystemSslRenewalConfigurationForm
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
        </SystemSslRenewalConfigurationForm>
    );
}
