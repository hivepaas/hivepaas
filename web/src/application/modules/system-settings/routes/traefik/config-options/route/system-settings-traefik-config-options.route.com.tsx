import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { TraefikConfigOptions_UpdateOne_Req } from "~/system-settings/api/services";
import { TraefikConfigOptionsCommands, TraefikConfigOptionsQueries } from "~/system-settings/data";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PageError } from "@application/shared/pages";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { TraefikConfigOptionsForm } from "../form";
import { mapFormOutputToUpdatePayload } from "../form/traefik-config-options.form-mappers";
import type { TraefikConfigOptionsFormOutput } from "../schemas";
import type { TraefikConfigOptionsFormRef } from "../types";

type UpdatePayload = TraefikConfigOptions_UpdateOne_Req["data"]["payload"];

function mapFormValuesToPayload(values: TraefikConfigOptionsFormOutput): UpdatePayload {
    return mapFormOutputToUpdatePayload(values);
}

export function SystemSettingsTraefikConfigOptionsRoute() {
    const formRef = useRef<TraefikConfigOptionsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const configOptionsQuery = TraefikConfigOptionsQueries.useFindOne();

    console.log("configOptionsQuery", configOptionsQuery);

    const { mutate: update, isPending } = TraefikConfigOptionsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Traefik config options updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update Traefik config options");
            }
        },
    });

    function handleSubmit(values: TraefikConfigOptionsFormOutput) {
        if (!canWrite) {
            return;
        }

        update({
            payload: mapFormValuesToPayload(values),
        });
    }

    if (configOptionsQuery.isLoading) {
        return <AppLoader />;
    }

    if (configOptionsQuery.error) {
        return (
            <PageError
                error={configOptionsQuery.error}
                onRetry={configOptionsQuery.refetch}
            />
        );
    }

    invariant(configOptionsQuery.data, "traefik config options data must be defined");

    return (
        <TraefikConfigOptionsForm
            ref={formRef}
            defaultValues={configOptionsQuery.data.data}
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
        </TraefikConfigOptionsForm>
    );
}
