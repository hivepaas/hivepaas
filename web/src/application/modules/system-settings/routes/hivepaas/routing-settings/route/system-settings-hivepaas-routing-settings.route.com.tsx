import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { HivePaaSRoutingSettings_UpdateOne_Req } from "~/system-settings/api/services";
import { HivePaaSRoutingSettingsCommands, HivePaaSRoutingSettingsQueries } from "~/system-settings/data";
import { useSettingsChangeConfirmDialogState } from "~/system-settings/dialogs";
import { probationWarning } from "~/system-settings/module-shared/utils";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { useGlobalAlertDialogState } from "@application/shared/dialogs";
import { PageError } from "@application/shared/pages";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { HivePaaSRoutingSettingsForm } from "../form";
import { mapFormValuesToPayload } from "../form/hivepaas-routing-settings.form-mappers";
import type { HivePaaSRoutingSettingsFormOutput } from "../schemas";
import type { HivePaaSRoutingSettingsFormRef } from "../types";

type UpdatePayload = HivePaaSRoutingSettings_UpdateOne_Req["data"]["payload"];

export function SystemSettingsHivePaaSRoutingSettingsRoute() {
    const formRef = useRef<HivePaaSRoutingSettingsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const confirmDialog = useSettingsChangeConfirmDialogState();
    const globalAlert = useGlobalAlertDialogState();
    const settingsQuery = HivePaaSRoutingSettingsQueries.useFindOne();

    const { mutate: update, isPending } = HivePaaSRoutingSettingsCommands.useUpdateOne({
        onSuccess: response => {
            toast.success("Routing settings applied - confirm to keep them");
            // Straight from the response rather than waiting for the refetch: the
            // countdown is already running server-side, and every second spent
            // waiting for a round trip is a second off the operator's budget.
            if (response.data.pendingChange != null) {
                confirmDialog.open("routing", response.data.pendingChange);
            }
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update HivePaaS Routing settings");
            }
        },
    });

    function handleSubmit(values: HivePaaSRoutingSettingsFormOutput) {
        if (!canWrite) {
            return;
        }

        const settings = settingsQuery.data?.data;
        invariant(settings, "hivepaas routing settings must be defined");

        const payload: UpdatePayload = mapFormValuesToPayload(values, settings.updateVer);

        // Every routing change goes on trial, so the warning is unconditional here.
        const { title, description } = probationWarning("routing settings");
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

    invariant(settingsQuery.data, "hivepaas routing settings data must be defined");

    return (
        <HivePaaSRoutingSettingsForm
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
        </HivePaaSRoutingSettingsForm>
    );
}

export { SystemSettingsHivePaaSRoutingSettingsRoute as SystemSettingsHivePaaSHttpSettingsRoute };
