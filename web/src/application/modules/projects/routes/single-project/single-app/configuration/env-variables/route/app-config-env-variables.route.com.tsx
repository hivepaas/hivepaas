import { useRef } from "react";

import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectAppEnvVarsCommands } from "~/projects/data/commands";
import { ProjectAppEnvVarsQueries } from "~/projects/data/queries";

import { AppLoader } from "@application/shared/components";
import { PageError } from "@application/shared/pages";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigEnvVarsForm } from "../form";
import { type AppConfigEnvVarsFormSchemaOutput } from "../schemas";
import { type AppConfigEnvVarsFormRef } from "../types";

export function AppConfigEnvVariablesRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const formRef = useRef<AppConfigEnvVarsFormRef>(null);

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const {
        data: envVarsData,
        isLoading: isLoadingEnvVars,
        error: envVarsError,
        refetch: refetchEnvVars,
    } = ProjectAppEnvVarsQueries.useFindOne({ projectID: projectId, appID: appId });

    const { mutate: update, isPending } = ProjectAppEnvVarsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Environment variables updated");
        },
        onError: (err: Error) => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    function handleSubmit(values: AppConfigEnvVarsFormSchemaOutput) {
        invariant(projectId, "projectId must be defined");
        invariant(appId, "appId must be defined");
        invariant(envVarsData, "envVarsData must be defined");

        update({
            projectID: projectId,
            appID: appId,
            ...values,
            updateVer: envVarsData.data.updateVer,
        });
    }

    if (isLoadingEnvVars) {
        return <AppLoader />;
    }

    if (envVarsError) {
        return (
            <PageError
                error={envVarsError}
                onRetry={refetchEnvVars}
            />
        );
    }

    invariant(envVarsData, "envVarsData must be defined");

    const { data: envVars } = envVarsData;

    return (
        <AppConfigEnvVarsForm
            ref={formRef}
            defaultValues={{
                buildtime: envVars.buildtime,
                runtime: envVars.runtime,
            }}
            onSubmit={handleSubmit}
        >
            <div className="flex justify-end gap-2 mt-4">
                <Button
                    type="submit"
                    className="min-w-[100px]"
                    disabled={isPending}
                    isLoading={isPending}
                >
                    Save
                </Button>
            </div>
        </AppConfigEnvVarsForm>
    );
}
