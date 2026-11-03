import { useRef } from "react";

import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppRoutingSettingsCommands, AppRoutingSettingsQueries } from "~/projects/data";
import { APP_CONFIGURATION_QUERY_OPTIONS } from "~/projects/data/constants";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";

import { AppLoader, FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigRoutingSettingsForm } from "../form";
import { mapFormValuesToPayload } from "../form/app-config-routing-settings.form-mappers";
import { type AppConfigRoutingSettingsFormSchemaOutput } from "../schemas";
import { type AppConfigRoutingSettingsFormRef } from "../types";

export function AppConfigRoutingSettingsRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();
    const formRef = useRef<AppConfigRoutingSettingsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading } = AppRoutingSettingsQueries.useFindOne(
        {
            projectID: projectId,
            env,
            appID: appId,
        },
        APP_CONFIGURATION_QUERY_OPTIONS,
    );

    const { mutate: update, isPending } = AppRoutingSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Routing settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update routing settings");
            }
        },
    });

    function handleSubmit(values: AppConfigRoutingSettingsFormSchemaOutput) {
        if (!canWrite) {
            return;
        }

        invariant(projectId, "projectId must be defined");
        invariant(env, "env must be defined");
        invariant(appId, "appId must be defined");

        const payload = mapFormValuesToPayload(values);
        payload.updateVer = data?.data.updateVer ?? 0;

        update({
            projectID: projectId,
            env,
            appID: appId,
            payload,
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <div className="flex flex-col gap-4">
            <AppConfigRoutingSettingsForm
                ref={formRef}
                defaultValues={data?.data}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
            >
                <FormActionBar>
                    <ProjectPermissionSubmitButton isPending={isPending} />
                </FormActionBar>
            </AppConfigRoutingSettingsForm>
        </div>
    );
}

export { AppConfigRoutingSettingsRoute as AppConfigHttpSettingsRoute };
