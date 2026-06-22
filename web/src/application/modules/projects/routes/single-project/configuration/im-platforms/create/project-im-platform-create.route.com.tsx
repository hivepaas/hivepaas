import { useParams } from "react-router";
import { ImPlatformFormRoute } from "~/settings/module-shared/components/im-platform-form-route";

export function ProjectImPlatformCreateRoute() {
    const { id: projectId = "" } = useParams();

    return (
        <ImPlatformFormRoute
            mode="create"
            scope={{ type: "project", projectId }}
        />
    );
}
