import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { HivePaaSServiceSettings_UpdateOne_Req } from "~/system-settings/api/services";
import { HivePaaSServiceSettingsCommands, HivePaaSServiceSettingsQueries } from "~/system-settings/data";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { HivePaaSGeneralForm } from "../form";
import type { HivePaaSGeneralFormOutput } from "../schemas";
import type { HivePaaSGeneralFormRef } from "../types";

type UpdatePayload = HivePaaSServiceSettings_UpdateOne_Req["data"]["payload"];

function mapFormValuesToPayload(values: HivePaaSGeneralFormOutput, updateVer: number): UpdatePayload {
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

export function SystemSettingsHivePaaSGeneralRoute() {
    const formRef = useRef<HivePaaSGeneralFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const settingsQuery = HivePaaSServiceSettingsQueries.useFindOne();

    const { mutate: update, isPending } = HivePaaSServiceSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("HivePaaS service settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update HivePaaS service settings");
            }
        },
    });

    function handleSubmit(values: HivePaaSGeneralFormOutput) {
        if (!canWrite) {
            return;
        }

        const settings = settingsQuery.data?.data;
        invariant(settings, "hivepaas service settings must be defined");

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

    invariant(settingsQuery.data, "hivepaas service settings data must be defined");

    return (
        <HivePaaSGeneralForm
            ref={formRef}
            defaultValues={settingsQuery.data.data}
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
        </HivePaaSGeneralForm>
    );
}
