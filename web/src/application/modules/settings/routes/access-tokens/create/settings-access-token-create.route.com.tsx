import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { AccessTokenFormRoute } from "~/settings/module-shared/components/access-token-form-route";

export function SettingsAccessTokenCreateRoute() {
    return (
        <div className={cn(listBox)}>
            <AccessTokenFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
