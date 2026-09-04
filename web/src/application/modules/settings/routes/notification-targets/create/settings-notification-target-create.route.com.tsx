import { formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { NotificationTargetFormRoute } from "~/settings/module-shared/components/notification-target-form-route";

export function SettingsNotificationTargetCreateRoute() {
    return (
        <div className={cn(formBox)}>
            <NotificationTargetFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
