import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { HivePaaSHttpSettings_UpdateOne_Req } from "~/system-settings/api/services";
import { HivePaaSHttpSettingsCommands, HivePaaSHttpSettingsQueries } from "~/system-settings/data";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { HivePaaSHttpSettingsForm } from "../form";
import { mapFormValuesToPayload } from "../form/hivepaas-http-settings.form-mappers";
import type { HivePaaSHttpSettingsFormOutput } from "../schemas";
import type { HivePaaSHttpSettingsFormRef } from "../types";

type UpdatePayload = HivePaaSHttpSettings_UpdateOne_Req["data"]["payload"];

export function SystemSettingsHivePaaSHttpSettingsRoute() {
    const formRef = useRef<HivePaaSHttpSettingsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const settingsQuery = HivePaaSHttpSettingsQueries.useFindOne();

    const { mutate: update, isPending } = HivePaaSHttpSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("HivePaaS HTTP settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update HivePaaS HTTP settings");
            }
        },
    });

    function handleSubmit(values: HivePaaSHttpSettingsFormOutput) {
        if (!canWrite) {
            return;
        }

        const settings = settingsQuery.data?.data;
        invariant(settings, "hivepaas http settings must be defined");

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

    invariant(settingsQuery.data, "hivepaas http settings data must be defined");

    return (
        <HivePaaSHttpSettingsForm
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
        </HivePaaSHttpSettingsForm>
    );
}
