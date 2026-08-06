import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { useProjectNotificationSettingsSources } from "~/projects/module-shared/hooks";

import { NotificationSettings } from "@application/shared/form";

import type { AppCloneSettingsFormSchemaInput } from "../schemas";

export function AppCloneNotificationFields({ readOnly = false }: Props) {
    const { id: projectId } = useParams<{ id: string }>();
    invariant(projectId, "projectId must be defined");

    const { sources, manageLink } = useProjectNotificationSettingsSources(projectId);

    return (
        <NotificationSettings<AppCloneSettingsFormSchemaInput>
            names={{
                successUseDefault: "notification.successUseDefault",
                success: "notification.success",
                failureUseDefault: "notification.failureUseDefault",
                failure: "notification.failure",
            }}
            sources={sources}
            manageLink={manageLink}
            readOnly={readOnly}
        />
    );
}

type Props = {
    readOnly?: boolean;
};
