import type { ModuleAction, ModulePermission } from "./module-permissions.types";

export type ProjectId = string;

export interface ProjectPermission<T extends ProjectId = ProjectId> {
    projectId: T;
    actions: ModulePermission["actions"];
}

export interface UseConditionalProjectParams<T extends ProjectId> {
    projectId: T;
}

export interface UseConditionalProjectResult<T extends ProjectId> {
    project: ProjectPermission<T> | null;
    actions: ProjectPermission["actions"];
    canRead: boolean;
    canWrite: boolean;
    canDelete: boolean;
    hasAccess: (action: ModuleAction) => boolean;
    source: "project" | "module" | "admin" | "none";
}
