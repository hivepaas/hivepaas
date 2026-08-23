import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { HivePaaSRoutingSettings_UpdateOne_Req } from "~/system-settings/api/services";
import { HivePaaSRoutingSettingsCommands, HivePaaSRoutingSettingsQueries } from "~/system-settings/data";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
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

    const settingsQuery = HivePaaSRoutingSettingsQueries.useFindOne();

    const { mutate: update, isPending } = HivePaaSRoutingSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("HivePaaS Routing settings updated");
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

        update({
            payload,
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
