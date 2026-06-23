import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { NotificationTargetFormRoute } from "~/settings/module-shared/components/notification-target-form-route";

export function SettingsNotificationTargetEditRoute() {
    const { notificationTargetId = "" } = useParams();

    return (
        <div className={cn(listBox)}>
            <NotificationTargetFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                notificationTargetId={notificationTargetId}
            />
        </div>
    );
}
