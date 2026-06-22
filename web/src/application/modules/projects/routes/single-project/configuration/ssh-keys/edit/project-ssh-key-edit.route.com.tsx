import { useParams } from "react-router";
import { SSHKeyFormRoute } from "~/settings/module-shared/components/ssh-key-form-route";

export function ProjectSSHKeyEditRoute() {
    const { sshKeyId = "", id: projectId = "" } = useParams();

    return (
        <SSHKeyFormRoute
            mode="edit"
            scope={{ type: "project", projectId }}
            sshKeyId={sshKeyId}
        />
    );
}
