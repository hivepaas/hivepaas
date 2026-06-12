import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import type { SystemBackup_UpdateOne_Req } from "~/system-settings/api/services";
import { SystemBackupCommands, SystemBackupQueries } from "~/system-settings/data";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { SystemBackupConfigurationForm } from "../form";
import { type SystemBackupConfigurationFormOutput, SystemBackupScheduleMode } from "../schemas";
import type { SystemBackupConfigurationFormRef } from "../types";

type UpdatePayload = SystemBackup_UpdateOne_Req["data"]["payload"];

function mapFormValuesToPayload(values: SystemBackupConfigurationFormOutput, updateVer: number): UpdatePayload {
    return {
        updateVer,
        status: values.status,
        schedule: {
            interval: values.scheduleMode === SystemBackupScheduleMode.Interval ? values.scheduleInterval : "",
            cronExpr: values.scheduleMode === SystemBackupScheduleMode.Cron ? values.scheduleCronExpr : "",
            initialTime: values.scheduleFrom ?? null,
        },
        compression: {
            format: values.compressionFormat,
        },
        encryption: {
            format: values.encryptionFormat,
            secret: values.encryptionSecret,
        },
        cloudStorage: {
            id: values.cloudStorage?.id ?? "",
            destinationDir: values.cloudStorageDestinationDir,
        },
        dbBackupConfig: {
            backupDeletedObjects: values.backupDeletedObjects,
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

export function SystemSettingsDataBackupConfigurationRoute() {
    const formRef = useRef<SystemBackupConfigurationFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const { data, isLoading } = SystemBackupQueries.useFindOne();

    const { mutate: update, isPending } = SystemBackupCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("System backup settings updated");
        },
    });

    function handleSubmit(values: SystemBackupConfigurationFormOutput) {
        if (!canWrite) {
            return;
        }

        update({
            payload: mapFormValuesToPayload(values, data?.data.updateVer ?? 0),
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <SystemBackupConfigurationForm
            ref={formRef}
            defaultValues={data?.data}
            onSubmit={handleSubmit}
            readOnly={!canWrite}
        >
            <div className="flex justify-end pt-4">
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
            </div>
        </SystemBackupConfigurationForm>
    );
}
