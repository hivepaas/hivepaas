import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { useProjectNotificationSettingsSources } from "~/projects/module-shared/hooks";

import { NotificationSettings } from "@application/shared/form";

import { type AppConfigDeploymentSettingsFormSchemaInput } from "../schemas";

export function NotificationFields({ readOnly = false }: Props) {
    const { id: projectId, env } = useParams<{ id: string; env: string }>();
    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");

    const { sources, manageLink } = useProjectNotificationSettingsSources(projectId, env);

    return (
        <NotificationSettings<AppConfigDeploymentSettingsFormSchemaInput>
            names={{
                successUseDefault: "notification.successUseDefault",
                success: "notification.success",
                failureUseDefault: "notification.failureUseDefault",
                failure: "notification.failure",
            }}
            sources={sources}
            manageLink={manageLink}
            readOnly={readOnly}
            titleWidth={220}
        />
    );
}

type Props = {
    readOnly?: boolean;
};
