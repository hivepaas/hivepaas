import { useParams } from "react-router";
import { NotificationTargetFormRoute } from "~/settings/module-shared/components/notification-target-form-route";

export function ProjectNotificationTargetCreateRoute() {
    const { id: projectId = "" } = useParams();

    return (
        <NotificationTargetFormRoute
            mode="create"
            scope={{ type: "project", projectId }}
        />
    );
}
