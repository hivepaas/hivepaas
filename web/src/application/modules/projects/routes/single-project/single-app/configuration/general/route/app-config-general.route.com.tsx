import { useRef } from "react";

import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { ProjectAppsCommands } from "~/projects/data/commands";
import { ProjectAppsQueries } from "~/projects/data/queries";

import { AppLoader } from "@application/shared/components";
import { PageError } from "@application/shared/pages";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigGeneralForm } from "../form";
import { type AppConfigGeneralFormSchemaOutput } from "../schemas";
import { type AppConfigGeneralFormRef } from "../types";

export function AppConfigGeneralRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const formRef = useRef<AppConfigGeneralFormRef>(null);

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading, error, refetch } = ProjectAppsQueries.useFindOneById({
        projectID: projectId,
        appID: appId,
    });

    const { mutate: update, isPending } = ProjectAppsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("App information updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            }
        },
    });

    function handleSubmit(values: AppConfigGeneralFormSchemaOutput) {
        invariant(projectId, "projectId must be defined");
        invariant(appId, "appId must be defined");
        invariant(data, "data must be defined");

        update({
            projectID: projectId,
            appID: appId,
            ...values,
            updateVer: data.data.updateVer,
            status: data.data.status,
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    if (error) {
        return (
            <PageError
                error={error}
                onRetry={refetch}
            />
        );
    }

    invariant(data, "data must be defined");

    const { data: app } = data;

    return (
        <AppConfigGeneralForm
            ref={formRef}
            defaultValues={app}
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
        </AppConfigGeneralForm>
    );
}
