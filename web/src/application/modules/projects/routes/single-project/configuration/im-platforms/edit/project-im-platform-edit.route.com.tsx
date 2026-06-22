import { useParams } from "react-router";
import { ImPlatformFormRoute } from "~/settings/module-shared/components/im-platform-form-route";

export function ProjectImPlatformEditRoute() {
    const { imPlatformId = "", id: projectId = "" } = useParams();

    return (
        <ImPlatformFormRoute
            mode="edit"
            scope={{ type: "project", projectId }}
            imPlatformId={imPlatformId}
        />
    );
}
