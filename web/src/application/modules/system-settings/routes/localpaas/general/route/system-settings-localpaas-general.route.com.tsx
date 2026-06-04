import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { LocalPaaSServiceSettings_UpdateOne_Req } from "~/system-settings/api/services";
import { LocalPaaSServiceSettingsCommands, LocalPaaSServiceSettingsQueries } from "~/system-settings/data";

import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { LocalPaaSGeneralForm } from "../form";
import type { LocalPaaSGeneralFormOutput } from "../schemas";
import type { LocalPaaSGeneralFormRef } from "../types";

type UpdatePayload = LocalPaaSServiceSettings_UpdateOne_Req["data"]["payload"];

function mapFormValuesToPayload(values: LocalPaaSGeneralFormOutput, updateVer: number): UpdatePayload {
    return {
        updateVer,
        appSettings: {
            replicas: values.appSettings.replicas,
        },
        workerSettings: {
            replicas: values.workerSettings.replicas,
            concurrency: values.workerSettings.concurrency,
            runWorkerInMainApp: values.workerSettings.runWorkerInMainApp,
        },
        taskSettings: {
            taskCheckInterval: values.taskSettings.taskCheckInterval,
            taskCreateInterval: values.taskSettings.taskCreateInterval,
        },
        healthcheckSettings: {
            baseInterval: values.healthcheckSettings.baseInterval,
        },
    };
}

export function SystemSettingsLocalPaaSGeneralRoute() {
    const formRef = useRef<LocalPaaSGeneralFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const settingsQuery = LocalPaaSServiceSettingsQueries.useFindOne();

    const { mutate: update, isPending } = LocalPaaSServiceSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("LocalPaaS service settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update LocalPaaS service settings");
            }
        },
    });

    function handleSubmit(values: LocalPaaSGeneralFormOutput) {
        if (!canWrite) {
            return;
        }

        const settings = settingsQuery.data?.data;
        invariant(settings, "localpaas service settings must be defined");

        update({
            payload: mapFormValuesToPayload(values, settings.updateVer),
        });
    }

    if (settingsQuery.isLoading) {
        return <AppLoader />;
    }

    if (settingsQuery.error) {
        return (
            <PageError
                error={settingsQuery.error}
                onRetry={settingsQuery.refetch}
            />
        );
    }

    invariant(settingsQuery.data, "localpaas service settings data must be defined");

    return (
        <LocalPaaSGeneralForm
            ref={formRef}
            defaultValues={settingsQuery.data.data}
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
        </LocalPaaSGeneralForm>
    );
}
