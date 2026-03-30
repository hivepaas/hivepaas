import { useRef } from "react";

import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppServiceSettingsCommands, AppServiceSettingsQueries } from "~/projects/data";

import { AppLoader } from "@application/shared/components";

import { EServiceMode } from "@application/modules/projects/module-shared/enums";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigAvailabilityForm } from "../form";
import { type AppConfigAvailabilitySchemaOutput } from "../schemas";
import { type AppConfigAvailabilityFormRef } from "../types";

export function AppConfigAvailabilityRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const formRef = useRef<AppConfigAvailabilityFormRef>(null);

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading } = AppServiceSettingsQueries.useFindOne({
        projectID: projectId,
        appID: appId,
    });

    const { mutate: update, isPending } = AppServiceSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Availability and scaling settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update availability and scaling settings");
            }
        },
    });

    function handleSubmit(values: AppConfigAvailabilitySchemaOutput) {
        invariant(projectId, "projectId must be defined");
        invariant(appId, "appId must be defined");

        update({
            projectID: projectId,
            appID: appId,
            payload: {
                ...values,
                updateVer: data?.data.updateVer ?? 0,
                modeSpec: {
                    mode: values.mode,
                    serviceReplicas: values.mode === EServiceMode.Replicated ? values.serviceReplicas : null,
                    jobMaxConcurrent: values.mode === EServiceMode.ReplicatedJob ? values.jobMaxConcurrent : null,
                    jobTotalCompletions: values.mode === EServiceMode.ReplicatedJob ? values.jobTotalCompletions : null,
                },
                placement: {
                    constraints: values.constraints,
                    preferences: values.preferences,
                },
            },
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <div className="flex flex-col gap-4">
            <div className="rounded-md border border-dashed border-primary bg-accent p-2 text-sm text-muted-foreground">
                For configuration details, see{" "}
                <a
                    href="https://docs.docker.com/reference/api/engine/version/v1.52/#tag/Service/operation/ServiceUpdate"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-500 underline underline-offset-2"
                >
                    docs
                </a>
            </div>

            <AppConfigAvailabilityForm
                ref={formRef}
                defaultValues={data?.data}
                onSubmit={handleSubmit}
            >
                <div className="flex justify-end mt-4">
                    <Button
                        type="submit"
                        className="min-w-[100px]"
                        disabled={isPending}
                        isLoading={isPending}
                    >
                        Save
                    </Button>
                </div>
            </AppConfigAvailabilityForm>
        </div>
    );
}
