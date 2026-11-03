import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { TraefikServiceSettings_UpdateOne_Req } from "~/system-settings/api/services";
import { TraefikServiceSettingsCommands, TraefikServiceSettingsQueries } from "~/system-settings/data";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { TraefikGeneralForm } from "../form";
import type { TraefikGeneralFormOutput } from "../schemas";
import type { TraefikGeneralFormRef } from "../types";

type UpdatePayload = TraefikServiceSettings_UpdateOne_Req["data"]["payload"];

function mapFormValuesToPayload(values: TraefikGeneralFormOutput, updateVer: number): UpdatePayload {
    return {
        updateVer,
        appSettings: {
            replicas: values.appSettings.replicas,
        },
    };
}

export function SystemSettingsTraefikGeneralRoute() {
    const formRef = useRef<TraefikGeneralFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const settingsQuery = TraefikServiceSettingsQueries.useFindOne();

    const { mutate: update, isPending } = TraefikServiceSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Traefik service settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update Traefik service settings");
            }
        },
    });

    function handleSubmit(values: TraefikGeneralFormOutput) {
        if (!canWrite) {
            return;
        }

        const settings = settingsQuery.data?.data;
        invariant(settings, "traefik service settings must be defined");

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

    invariant(settingsQuery.data, "traefik service settings data must be defined");

    return (
        <TraefikGeneralForm
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
        </TraefikGeneralForm>
    );
}
