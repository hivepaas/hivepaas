import { useProjectPermissionsStore } from "../store";
import type { ProjectId, ProjectPermission } from "../types";

interface UseSetProjectPermissionsResult {
    setProjectPermissions: (projects: readonly ProjectPermission[]) => void;
    upsertProjectPermission: (project: ProjectPermission) => void;
    removeProjectPermission: (projectId: ProjectId) => void;
    clearProjectPermissions: () => void;
}

export function useSetProjectPermissions(): UseSetProjectPermissionsResult {
    const setProjects = useProjectPermissionsStore(state => state.setProjects);
    const upsertProject = useProjectPermissionsStore(state => state.upsertProject);
    const removeProject = useProjectPermissionsStore(state => state.removeProject);
    const clearProjects = useProjectPermissionsStore(state => state.clearProjects);

    return {
        setProjectPermissions: setProjects,
        upsertProjectPermission: upsertProject,
        removeProjectPermission: removeProject,
        clearProjectPermissions: clearProjects,
    };
}
