import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SSHKeyFormRoute } from "~/settings/module-shared/components/ssh-key-form-route";

export function SettingsSSHKeyCreateRoute() {
    return (
        <div className={cn(listBox)}>
            <SSHKeyFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
