import { useRef } from "react";

import { Button } from "@components/ui";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import type { TraefikConfigOptions_UpdateOne_Req } from "~/system-settings/api/services";
import { TraefikConfigOptionsCommands, TraefikConfigOptionsQueries } from "~/system-settings/data";
import { useSettingsChangeConfirmDialogState } from "~/system-settings/dialogs";
import { probationWarning } from "~/system-settings/module-shared/utils";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { useGlobalAlertDialogState } from "@application/shared/dialogs";
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

    const confirmDialog = useSettingsChangeConfirmDialogState();
    const globalAlert = useGlobalAlertDialogState();
    const configOptionsQuery = TraefikConfigOptionsQueries.useFindOne();

    const { mutate: update, isPending } = TraefikConfigOptionsCommands.useUpdateOne({
        onSuccess: response => {
            // No trial means the request asked for nothing Traefik was not already
            // running, so nothing restarted and there is nothing to confirm.
            if (response.data.pendingChange == null) {
                toast.success("Traefik config options updated");
                return;
            }

            toast.success("Traefik config applied - confirm to keep it");
            // Straight from the response rather than waiting for the refetch: the
            // countdown is already running server-side, and every second spent
            // waiting for a round trip is a second off the operator's budget.
            confirmDialog.open("traefik", response.data.pendingChange);
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

        const payload = mapFormValuesToPayload(values);

        // Unconditional, because the cost is not in whether the change is risky
        // but in what applying it does at all: Traefik's task is replaced, and its
        // ports are bound in host mode, so every route is down until the new one
        // is serving. Nobody should meet that by accident.
        const { title, description } = probationWarning("Traefik config options");
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
