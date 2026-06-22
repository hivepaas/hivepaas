import { useParams } from "react-router";
import { NotificationTargetFormRoute } from "~/settings/module-shared/components/notification-target-form-route";

export function ProjectNotificationTargetEditRoute() {
    const { notificationTargetId = "", id: projectId = "" } = useParams();

    return (
        <NotificationTargetFormRoute
            mode="edit"
            scope={{ type: "project", projectId }}
            notificationTargetId={notificationTargetId}
        />
    );
}
