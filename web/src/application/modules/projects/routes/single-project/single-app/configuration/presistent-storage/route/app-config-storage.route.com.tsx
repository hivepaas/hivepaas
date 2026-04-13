import { useRef } from "react";

import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppStorageSettingsCommands, AppStorageSettingsQueries } from "~/projects/data";

import { AppLoader } from "@application/shared/components";

import { isValidationException } from "@infrastructure/api";
import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigStorageForm } from "../form";
import { mapFormValuesToAppStoragePayload } from "../form/app-config-storage.form-mappers";
import { type AppConfigStorageFormSchemaOutput } from "../schemas";
import { type AppConfigStorageFormRef } from "../types";

export function AppConfigStorageRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const formRef = useRef<AppConfigStorageFormRef>(null);

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading } = AppStorageSettingsQueries.useFindOne({
        projectID: projectId,
        appID: appId,
    });

    const { mutate: update, isPending } = AppStorageSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Storage settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update storage settings");
            }
        },
    });

    function handleSubmit(values: AppConfigStorageFormSchemaOutput) {
        invariant(projectId, "projectId must be defined");
        invariant(appId, "appId must be defined");

        update({
            projectID: projectId,
            appID: appId,
            payload: mapFormValuesToAppStoragePayload(values, data?.data),
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
                .
            </div>

            <AppConfigStorageForm
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
            </AppConfigStorageForm>
        </div>
    );
}
