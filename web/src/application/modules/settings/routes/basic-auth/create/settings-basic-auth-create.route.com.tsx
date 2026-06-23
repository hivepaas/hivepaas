import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { BasicAuthFormRoute } from "~/settings/module-shared/components/basic-auth-form-route";

export function SettingsBasicAuthCreateRoute() {
    return (
        <div className={cn(listBox)}>
            <BasicAuthFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
