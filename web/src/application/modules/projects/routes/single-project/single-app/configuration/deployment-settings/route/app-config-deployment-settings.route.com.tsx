import { useRef } from "react";

import { Button } from "@components/ui";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { type AppDeploymentSettings_UpdateOne_Req } from "~/projects/api/services";
import { AppDeploymentSettingsCommands, AppDeploymentSettingsQueries } from "~/projects/data";
import { EAppDeploymentMethod } from "~/projects/module-shared/enums";

import { AppLoader } from "@application/shared/components";

import { isValidationException } from "@infrastructure/api";

import { ValidationException } from "@infrastructure/exceptions/validation";

import { AppConfigDeploymentSettingsForm } from "../form";
import { type AppConfigDeploymentSettingsFormSchemaOutput } from "../schemas";
import { type AppConfigDeploymentSettingsFormRef } from "../types";

type DeploymentSettingsUpdatePayload = AppDeploymentSettings_UpdateOne_Req["data"]["payload"];

function buildNotificationPayload(
    notification: AppConfigDeploymentSettingsFormSchemaOutput["notification"],
): DeploymentSettingsUpdatePayload["notification"] {
    if (!notification) {
        return {
            successUseDefault: false,
            failureUseDefault: false,
        };
    }
    return {
        successUseDefault: notification.successUseDefault,
        ...(notification.success?.id ? { success: { id: notification.success.id } } : {}),
        failureUseDefault: notification.failureUseDefault,
        ...(notification.failure?.id ? { failure: { id: notification.failure.id } } : {}),
    };
}

function mapFormValuesToPayload(values: AppConfigDeploymentSettingsFormSchemaOutput): DeploymentSettingsUpdatePayload {
    const base = {
        command: values.command ?? "",
        workingDir: values.workingDir ?? "",
        preDeploymentCommand: values.preDeploymentCommand ?? "",
        postDeploymentCommand: values.postDeploymentCommand ?? "",
        notification: buildNotificationPayload(values.notification),
    };

    if (values.activeMethod === EAppDeploymentMethod.Repo) {
        return {
            ...base,
            activeMethod: values.activeMethod,
            repoSource: {
                buildTool: values.repoSource.buildTool,
                repoType: values.repoSource.repoType,
                repoUrl: values.repoSource.repoUrl,
                repoRef: values.repoSource.repoRef,
                credentials: { id: values.repoSource.credentials?.id ?? "" },
                imageName: values.repoSource.imageName ?? "",
            },
        };
    }

    // Contract marks `repoSource` required on payload; Image flow only sends `imageSource`.
    return {
        ...base,
        activeMethod: values.activeMethod,
        imageSource: {
            image: values.imageSource.image,
            registryAuth: { id: values.imageSource.registryAuth?.id ?? "" },
        },
    } as DeploymentSettingsUpdatePayload;
}

export function AppConfigDeploymentSettingsRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const formRef = useRef<AppConfigDeploymentSettingsFormRef>(null);

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isLoading } = AppDeploymentSettingsQueries.useFindOne({
        projectID: projectId,
        appID: appId,
    });

    const { mutate: update, isPending } = AppDeploymentSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Deployment settings updated");
        },
        onError: err => {
            if (isValidationException(err)) {
                formRef.current?.onError(ValidationException.fromHttp(err));
            } else if (err instanceof Error) {
                toast.error(err.message);
            } else {
                toast.error("Failed to update deployment settings");
            }
        },
    });

    function handleSubmit(values: AppConfigDeploymentSettingsFormSchemaOutput) {
        invariant(projectId, "projectId must be defined");
        invariant(appId, "appId must be defined");

        update({
            projectID: projectId,
            appID: appId,
            updateVer: data?.data.updateVer ?? 0,
            payload: mapFormValuesToPayload(values),
        });
    }

    if (isLoading) {
        return <AppLoader />;
    }

    return (
        <div className="flex flex-col gap-4">
            <AppConfigDeploymentSettingsForm
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
                        Deploy
                    </Button>
                </div>
            </AppConfigDeploymentSettingsForm>
        </div>
    );
}
