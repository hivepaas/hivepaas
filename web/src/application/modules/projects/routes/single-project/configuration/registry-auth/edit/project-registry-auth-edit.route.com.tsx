import { useParams } from "react-router";
import { RegistryAuthFormRoute } from "~/settings/module-shared/components/registry-auth-form-route";

export function ProjectRegistryAuthEditRoute() {
    const { registryAuthId = "", id: projectId = "" } = useParams();

    return (
        <RegistryAuthFormRoute
            mode="edit"
            scope={{ type: "project", projectId }}
            registryAuthId={registryAuthId}
        />
    );
}
