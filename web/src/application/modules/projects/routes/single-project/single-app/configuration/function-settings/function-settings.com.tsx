import { useRef } from "react";

import { toast } from "sonner";
import { AppDeploymentSettingsCommands } from "~/projects/data";
import { type FunctionMethod } from "~/projects/domain";
import { ProjectPermissionSubmitButton } from "~/projects/module-shared/components";
import { functionSettingsToPayload } from "~/projects/module-shared/utils";

import { FormActionBar } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { FunctionSettingsForm, type FunctionSettingsFormRef } from "./form";
import { type FunctionSettingsFormOutput, functionSettingsToSource } from "./schemas";

/**
 * A function's deployment settings: its runtime, its limits, where its code is.
 * Saving them deploys the function.
 */
export function FunctionSettings({ projectId, env, appId, settings }: Props) {
    const formRef = useRef<FunctionSettingsFormRef>(null);
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const { mutate: update, isPending } = AppDeploymentSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Function settings saved, a deployment has started");
        },
        // The failure itself is notified by the API's hook: this puts it on its fields.
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    function handleSubmit(values: FunctionSettingsFormOutput) {
        if (!canWrite) {
            return;
        }

        update({
            projectID: projectId,
            env,
            appID: appId,
            updateVer: settings.updateVer,
            payload: functionSettingsToPayload(settings, functionSettingsToSource(values, settings.functionSource)),
        });
    }

    return (
        <div className="flex flex-col gap-4">
            <FunctionSettingsForm
                // Saved settings come back with a new version: the form starts again from them.
                key={settings.updateVer}
                ref={formRef}
                projectId={projectId}
                env={env}
                source={settings.functionSource}
                onSubmit={handleSubmit}
                readOnly={!canWrite}
            >
                <FormActionBar>
                    <ProjectPermissionSubmitButton
                        isPending={isPending}
                        label="Save & Deploy"
                    />
                </FormActionBar>
            </FunctionSettingsForm>
        </div>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    settings: FunctionMethod;
}
