import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { RegistryAuthFormRoute } from "~/settings/module-shared/components/registry-auth-form-route";

export function SettingsRegistryAuthCreateRoute() {
    return (
        <div className={cn(listBox)}>
            <RegistryAuthFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
