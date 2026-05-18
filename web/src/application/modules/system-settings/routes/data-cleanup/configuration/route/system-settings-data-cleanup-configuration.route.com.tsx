import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import type { SystemCleanup_UpdateOne_Req } from "~/system-settings/api/services";
import { SystemCleanupCommands, SystemCleanupQueries } from "~/system-settings/data";

import { AppLoader } from "@application/shared/components";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { SystemCleanupConfigurationForm } from "../form";
import type { SystemCleanupConfigurationFormOutput } from "../schemas";
import type { SystemCleanupConfigurationFormRef } from "../types";

type UpdatePayload = SystemCleanup_UpdateOne_Req["data"]["payload"];

function mapFormValuesToPayload(values: SystemCleanupConfigurationFormOutput, updateVer: number): UpdatePayload {
    return {
        updateVer,
        status: values.status,
        scheduleInterval: values.scheduleInterval,
        scheduleFrom: values.scheduleFrom,
        dbObjectRetention: {
            enabled: values.dbObjectRetention.enabled,
            tasks: values.dbObjectRetention.tasks,
            deployments: values.dbObjectRetention.deployments,
            sysErrors: values.dbObjectRetention.sysErrors,
            deletedObjects: values.dbObjectRetention.deletedObjects,
        },
        clusterCleanup: {
            enabled: values.clusterCleanup.enabled,
            pruneImages: values.clusterCleanup.pruneImages,
            pruneVolumes: values.clusterCleanup.pruneVolumes,
            pruneNetworks: values.clusterCleanup.pruneNetworks,
            pruneContainers: values.clusterCleanup.pruneContainers,
        },
        backupCleanup: {
            enabled: values.backupCleanup.enabled,
            localBackupRetention: values.backupCleanup.localBackupRetention,
            cloudBackupRetention: values.backupCleanup.cloudBackupRetention,
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

export function SystemSettingsDataCleanupConfigurationRoute() {
    const formRef = useRef<SystemCleanupConfigurationFormRef>(null);

    const { data, isLoading } = SystemCleanupQueries.useFindOne();

    const { mutate: update, isPending } = SystemCleanupCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("System cleanup settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update system cleanup settings");
            }
        },
    });

    function handleSubmit(values: SystemCleanupConfigurationFormOutput) {
        update({
            payload: mapFormValuesToPayload(values, data?.data.updateVer ?? 0),
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <SystemCleanupConfigurationForm
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
        </SystemCleanupConfigurationForm>
    );
}
