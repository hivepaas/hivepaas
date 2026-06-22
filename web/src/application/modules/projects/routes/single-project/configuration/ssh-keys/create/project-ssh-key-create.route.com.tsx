import { useParams } from "react-router";
import { SSHKeyFormRoute } from "~/settings/module-shared/components/ssh-key-form-route";

export function ProjectSSHKeyCreateRoute() {
    const { id: projectId = "" } = useParams();

    return (
        <SSHKeyFormRoute
            mode="create"
            scope={{ type: "project", projectId }}
        />
    );
}
