import { formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { KeyAuthFormRoute } from "~/settings/module-shared/components/key-auth-form-route";

export function SettingsKeyAuthCreateRoute() {
    return (
        <div className={cn(formBox)}>
            <KeyAuthFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
