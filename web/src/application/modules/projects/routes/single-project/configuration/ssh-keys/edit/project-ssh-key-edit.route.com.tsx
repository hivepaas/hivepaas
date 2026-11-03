import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { SSHKeyFormRoute } from "~/settings/module-shared/components/ssh-key-form-route";

export function ProjectSSHKeyEditRoute() {
    const { id: projectId = "", sshKeyId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <SSHKeyFormRoute
            mode="edit"
            scope={{ type: "project", projectId, env: scopedEnv }}
            sshKeyId={sshKeyId}
        />
    );
}
