import { useEffect, useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { HivePaaSServiceSettings_UpdateOne_Req } from "~/system-settings/api/services";
import { HivePaaSServiceSettingsCommands, HivePaaSServiceSettingsQueries } from "~/system-settings/data";
import type { SettingsChangeOutcome } from "~/system-settings/dialogs";
import { useSettingsChangeConfirmDialogState } from "~/system-settings/dialogs";
import type { SettingsPendingChange } from "~/system-settings/domain";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { HivePaaSGeneralForm } from "../form";
import { splitTrustedIPsText } from "../hivepaas-general.constants";
import type { HivePaaSGeneralFormOutput } from "../schemas";
import type { HivePaaSGeneralFormRef } from "../types";

type UpdatePayload = HivePaaSServiceSettings_UpdateOne_Req["data"]["payload"];

/**
 * How long the dashboard asks to keep a proxy settings change on trial.
 *
 * Longer than the routing equivalent because this change restarts traefik: the
 * first minute or so of the countdown is spent with HivePaaS unreachable, before
 * the operator can do anything at all. The server clamps it, and its floor
 * already accounts for that restart - this only has to be generous.
 */
const CONFIRM_WINDOW = "5m";

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

export function SystemSettingsHivePaaSGeneralRoute() {
    const formRef = useRef<HivePaaSGeneralFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const confirmDialog = useSettingsChangeConfirmDialogState();
    // Changes this page has already seen through to an end. Without it the effect
    // below would reopen the dialog from cached data in the moment between a
    // confirmation landing and the refetch that clears the pending change.
    const resolvedChangeIds = useRef(new Set<string>());

    const settingsQuery = HivePaaSServiceSettingsQueries.useFindOne();

    const pendingChange = settingsQuery.data?.data.pendingChange ?? null;

    function openConfirmDialog(change: SettingsPendingChange) {
        if (resolvedChangeIds.current.has(change.changeId)) {
            return;
        }
        confirmDialog.open("service", change, {
            props: {
                onResolved: (_outcome: SettingsChangeOutcome, changeId: string) => {
                    resolvedChangeIds.current.add(changeId);
                    void settingsQuery.refetch();
                },
            },
        });
    }

    // Reopen the trial after a reload. This change restarts traefik, so the tab
    // that started it is quite likely to have been reloaded by the operator while
    // they waited - and losing the countdown would leave them watching a change
    // silently roll back.
    useEffect(() => {
        if (pendingChange != null) {
            openConfirmDialog(pendingChange);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [pendingChange]);

    const { mutate: update, isPending } = HivePaaSServiceSettingsCommands.useUpdateOne({
        onSuccess: response => {
            toast.success("HivePaaS service settings updated");
            // Only proxy changes go on trial; everything else answers with null.
            if (response.data.pendingChange != null) {
                openConfirmDialog(response.data.pendingChange);
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
