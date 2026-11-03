import { useMemo } from "react";

import { useProjectPermissionsStore } from "../store";
import type { ProjectId, ProjectPermission } from "../types";

interface UseConditionalProjectCollectionsResult {
    list: readonly ProjectPermission[];
    map: ReadonlyMap<ProjectId, ProjectPermission>;
}

export function useConditionalProjectCollections(): UseConditionalProjectCollectionsResult {
    const projects = useProjectPermissionsStore(state => state.projects);

    return useMemo(() => {
        const map: ReadonlyMap<ProjectId, ProjectPermission> = new Map(
            projects.map(project => [project.projectId, project]),
        );

        return {
            list: projects,
            map,
        };
    }, [projects]);
}
