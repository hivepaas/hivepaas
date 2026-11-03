import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { HivePaaSServiceSettings_UpdateOne_Req } from "~/system-settings/api/services";
import { HivePaaSServiceSettingsCommands, HivePaaSServiceSettingsQueries } from "~/system-settings/data";
import { useSettingsChangeConfirmDialogState } from "~/system-settings/dialogs";
import type { HivePaaSProxySettings } from "~/system-settings/domain";
import { CONFIRM_WINDOW, probationWarning } from "~/system-settings/module-shared/utils";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { useGlobalAlertDialogState } from "@application/shared/dialogs";
import { PageError } from "@application/shared/pages";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { HivePaaSGeneralForm } from "../form";
import { splitTrustedIPsText } from "../hivepaas-general.constants";
import type { HivePaaSGeneralFormOutput } from "../schemas";
import type { HivePaaSGeneralFormRef } from "../types";

type UpdatePayload = HivePaaSServiceSettings_UpdateOne_Req["data"]["payload"];

function mapFormValuesToPayload(values: HivePaaSGeneralFormOutput, updateVer: number): UpdatePayload {
    const proxyProvider = values.proxySettings.proxyProvider.trim();

    return {
        updateVer,
        confirmWindow: CONFIRM_WINDOW,
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
        periodicSettings: {
            baseInterval: values.periodicSettings.baseInterval,
            batchSize: values.periodicSettings.batchSize,
        },
        proxySettings: {
            proxyProvider,
            trustedIPs: proxyProvider ? splitTrustedIPsText(values.proxySettings.trustedIPsText) : [],
            proxyHops: proxyProvider ? values.proxySettings.proxyHops : 0,
        },
    };
}

/**
 * Mirrors HivePaaSProxySettings.Equal on the server, which is what actually
 * decides whether a change goes on trial.
 *
 * Being wrong in the permissive direction only costs a warning nobody needed;
 * being wrong the other way would let a change go on trial unannounced. The
 * server remains the authority - this only chooses whether to ask first.
 */
function proxySettingsChanged(next: HivePaaSProxySettings, current: HivePaaSProxySettings): boolean {
    return (
        next.proxyProvider !== current.proxyProvider ||
        next.proxyHops !== current.proxyHops ||
        next.trustedIPs.join(",") !== current.trustedIPs.join(",")
    );
}

export function SystemSettingsHivePaaSGeneralRoute() {
    const formRef = useRef<HivePaaSGeneralFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const confirmDialog = useSettingsChangeConfirmDialogState();
    const globalAlert = useGlobalAlertDialogState();
    const settingsQuery = HivePaaSServiceSettingsQueries.useFindOne();

    const { mutate: update, isPending } = HivePaaSServiceSettingsCommands.useUpdateOne({
        onSuccess: response => {
            toast.success("HivePaaS service settings updated");
            // Only proxy changes go on trial; everything else answers with null.
            if (response.data.pendingChange != null) {
                confirmDialog.open("service", response.data.pendingChange);
            }
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

        const payload = mapFormValuesToPayload(values, settings.updateVer);

        // Only a proxy change goes on trial, so only a proxy change is worth
        // warning about - nagging on every replica tweak would train people to
        // click through the one warning that matters.
        if (!proxySettingsChanged(payload.proxySettings, settings.proxySettings)) {
            update({ payload });
            return;
        }

        const { title, description } = probationWarning("proxy settings");
        globalAlert.open({
            props: {
                type: "warning",
                title,
                description,
                actionText: "Apply",
                onAction: () => {
                    update({ payload });
                },
            },
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
