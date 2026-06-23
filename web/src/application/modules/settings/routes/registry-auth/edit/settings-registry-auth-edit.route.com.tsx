import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { RegistryAuthFormRoute } from "~/settings/module-shared/components/registry-auth-form-route";

export function SettingsRegistryAuthEditRoute() {
    const { registryAuthId = "" } = useParams();

    return (
        <div className={cn(listBox)}>
            <RegistryAuthFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                registryAuthId={registryAuthId}
            />
        </div>
    );
}
