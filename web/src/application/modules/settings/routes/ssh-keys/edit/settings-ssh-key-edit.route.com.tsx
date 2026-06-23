import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { SSHKeyFormRoute } from "~/settings/module-shared/components/ssh-key-form-route";

export function SettingsSSHKeyEditRoute() {
    const { sshKeyId = "" } = useParams();

    return (
        <div className={cn(listBox)}>
            <SSHKeyFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                sshKeyId={sshKeyId}
            />
        </div>
    );
}
