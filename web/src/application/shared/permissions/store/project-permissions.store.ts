import { create } from "zustand";

import type { ProjectId, ProjectPermission } from "../types";

interface ProjectPermissionsState {
    projects: readonly ProjectPermission[];
    setProjects: (projects: readonly ProjectPermission[]) => void;
    upsertProject: (project: ProjectPermission) => void;
    removeProject: (projectId: ProjectId) => void;
    clearProjects: () => void;
}

export const useProjectPermissionsStore = create<ProjectPermissionsState>()(set => ({
    projects: [],

    setProjects: projects => {
        set({
            projects: [...projects],
        });
    },

    upsertProject: project => {
        set(state => {
            const next = state.projects.filter(item => item.projectId !== project.projectId);

            return {
                projects: [...next, project],
            };
        });
    },

    removeProject: projectId => {
        set(state => ({
            projects: state.projects.filter(project => project.projectId !== projectId),
        }));
    },

    clearProjects: () => {
        set({
            projects: [],
        });
    },
}));
