import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import type { SystemBackup_UpdateOne_Req } from "~/system-settings/api/services";
import { SystemBackupCommands, SystemBackupQueries } from "~/system-settings/data";

import { AppLoader } from "@application/shared/components";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { SystemBackupConfigurationForm } from "../form";
import type { SystemBackupConfigurationFormOutput } from "../schemas";
import type { SystemBackupConfigurationFormRef } from "../types";

type UpdatePayload = SystemBackup_UpdateOne_Req["data"]["payload"];

function mapFormValuesToPayload(values: SystemBackupConfigurationFormOutput, updateVer: number): UpdatePayload {
    return {
        updateVer,
        status: values.status,
        scheduleInterval: values.scheduleInterval,
        scheduleFrom: values.scheduleFrom,
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
                id: values.notification.success?.id ?? "",
            },
            failureUseDefault: values.notification.failureUseDefault,
            failure: {
                id: values.notification.failure?.id ?? "",
            },
        },
    };
}

export function SystemSettingsDataBackupConfigurationRoute() {
    const formRef = useRef<SystemBackupConfigurationFormRef>(null);

    const { data, isLoading } = SystemBackupQueries.useFindOne();

    const { mutate: update, isPending } = SystemBackupCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("System backup settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update system backup settings");
            }
        },
    });

    function handleSubmit(values: SystemBackupConfigurationFormOutput) {
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
        >
            <div className="flex justify-end pt-4">
                <Button
                    type="submit"
                    className="min-w-[100px]"
                    disabled={isPending}
                    isLoading={isPending}
                >
                    Save
                </Button>
            </div>
        </SystemBackupConfigurationForm>
    );
}
