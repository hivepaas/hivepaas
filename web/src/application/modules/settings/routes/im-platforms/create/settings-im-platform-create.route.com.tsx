import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { ImPlatformFormRoute } from "~/settings/module-shared/components/im-platform-form-route";

export function SettingsImPlatformCreateRoute() {
    return (
        <div className={cn(listBox)}>
            <ImPlatformFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
