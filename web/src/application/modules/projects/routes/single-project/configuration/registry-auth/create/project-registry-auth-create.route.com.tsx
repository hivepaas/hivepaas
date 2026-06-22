import { useParams } from "react-router";
import { RegistryAuthFormRoute } from "~/settings/module-shared/components/registry-auth-form-route";

export function ProjectRegistryAuthCreateRoute() {
    const { id: projectId = "" } = useParams();

    return (
        <RegistryAuthFormRoute
            mode="create"
            scope={{ type: "project", projectId }}
        />
    );
}
