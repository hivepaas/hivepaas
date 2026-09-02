import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import type { SystemBackupRepoCleanup_UpdateOne_Req } from "~/system-settings/api/services";
import { SystemBackupRepoCleanupCommands, SystemBackupRepoCleanupQueries } from "~/system-settings/data";
import type { SystemBackupRepoCleanupSettings } from "~/system-settings/domain";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { SystemBackupRepoCleanupConfigurationForm } from "../form";
import { type SystemBackupRepoCleanupConfigurationFormOutput, SystemBackupRepoCleanupScheduleMode } from "../schemas";
import type { SystemBackupRepoCleanupConfigurationFormRef } from "../types";

type UpdatePayload = SystemBackupRepoCleanup_UpdateOne_Req["data"]["payload"];

function mapFormValuesToPayload(
    values: SystemBackupRepoCleanupConfigurationFormOutput,
    settings?: SystemBackupRepoCleanupSettings,
): UpdatePayload {
    return {
        updateVer: settings?.updateVer ?? 0,
        status: values.status,
        schedule: {
            interval:
                values.scheduleMode === SystemBackupRepoCleanupScheduleMode.Interval ? values.scheduleInterval : "",
            cronExpr: values.scheduleMode === SystemBackupRepoCleanupScheduleMode.Cron ? values.scheduleCronExpr : "",
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

export function SystemSettingsBackupRepoCleanupConfigurationRoute() {
    const formRef = useRef<SystemBackupRepoCleanupConfigurationFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const { data, isLoading } = SystemBackupRepoCleanupQueries.useFindOne();

    const { mutate: update, isPending } = SystemBackupRepoCleanupCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Backup repo cleanup settings updated");
        },
    });

    function handleSubmit(values: SystemBackupRepoCleanupConfigurationFormOutput) {
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
        <SystemBackupRepoCleanupConfigurationForm
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
        </SystemBackupRepoCleanupConfigurationForm>
    );
}
