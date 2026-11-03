import { formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { SSHKeyFormRoute } from "~/settings/module-shared/components/ssh-key-form-route";

export function SettingsSSHKeyEditRoute() {
    const { sshKeyId = "" } = useParams();

    return (
        <div className={cn(formBox)}>
            <SSHKeyFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                sshKeyId={sshKeyId}
            />
        </div>
    );
}
