import { useRef } from "react";

import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppHttpSettingsCommands, AppHttpSettingsQueries } from "~/projects/data";

import { AppLoader } from "@application/shared/components";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigHttpSettingsForm } from "../form";
import { mapFormValuesToPayload } from "../form/app-config-http-settings.form-mappers";
import { type AppConfigHttpSettingsFormSchemaOutput } from "../schemas";
import { type AppConfigHttpSettingsFormRef } from "../types";

export function AppConfigHttpSettingsRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const formRef = useRef<AppConfigHttpSettingsFormRef>(null);

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading } = AppHttpSettingsQueries.useFindOne({
        projectID: projectId,
        appID: appId,
    });

    const { mutate: update, isPending } = AppHttpSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("HTTP settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update HTTP settings");
            }
        },
    });

    function handleSubmit(values: AppConfigHttpSettingsFormSchemaOutput) {
        invariant(projectId, "projectId must be defined");
        invariant(appId, "appId must be defined");

        const payload = mapFormValuesToPayload(values);
        payload.updateVer = data?.data.updateVer ?? 0;

        update({
            projectID: projectId,
            appID: appId,
            payload,
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <div className="flex flex-col gap-4">
            <AppConfigHttpSettingsForm
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
            </AppConfigHttpSettingsForm>
        </div>
    );
}
