import { create } from "zustand";

export const PROJECT_ENV_FILTER_ALL = "all";

const STORAGE_KEY = "project-env-filter";

type ProjectEnvSelections = Record<string, string>;

class ProjectEnvFilterStorage {
    persisted(): ProjectEnvSelections {
        if (typeof window === "undefined") {
            return {};
        }

        const value = window.localStorage.getItem(STORAGE_KEY);

        if (!value) {
            return {};
        }

        try {
            const json = JSON.parse(value) as unknown;

            if (!json || typeof json !== "object" || Array.isArray(json)) {
                window.localStorage.removeItem(STORAGE_KEY);
                return {};
            }

            return Object.entries(json).reduce<ProjectEnvSelections>((acc, [projectId, env]) => {
                if (typeof env !== "string") {
                    return acc;
                }

                return {
                    ...acc,
                    [projectId]: env,
                };
            }, {});
        } catch {
            window.localStorage.removeItem(STORAGE_KEY);
            return {};
        }
    }

    update(data: ProjectEnvSelections): void {
        if (typeof window === "undefined") {
            return;
        }

        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
    }
}

interface State {
    selectedEnvs: ProjectEnvSelections;
}

interface Actions {
    setSelectedEnv: (projectId: string, env: string) => void;
    normalizeSelectedEnv: (projectId: string, envNames: string[]) => void;
}

const storage = new ProjectEnvFilterStorage();

export const useProjectEnvFilterStore = create<State & Actions>()(set => ({
    selectedEnvs: storage.persisted(),

    setSelectedEnv: (projectId, env) => {
        if (!projectId) {
            return;
        }

        set(state => {
            const selectedEnvs = {
                ...state.selectedEnvs,
                [projectId]: env,
            };

            storage.update(selectedEnvs);

            return { selectedEnvs };
        });
    },

    normalizeSelectedEnv: (projectId, envNames) => {
        if (!projectId) {
            return;
        }

        set(state => {
            const selectedEnv = state.selectedEnvs[projectId] ?? PROJECT_ENV_FILTER_ALL;

            if (selectedEnv === PROJECT_ENV_FILTER_ALL || envNames.includes(selectedEnv)) {
                return {};
            }

            const selectedEnvs = {
                ...state.selectedEnvs,
                [projectId]: PROJECT_ENV_FILTER_ALL,
            };

            storage.update(selectedEnvs);

            return { selectedEnvs };
        });
    },
}));

export function useSelectedProjectEnv(projectId: string): string {
    return useProjectEnvFilterStore(state => state.selectedEnvs[projectId] ?? PROJECT_ENV_FILTER_ALL);
}

export function useProjectEnvFilter(projectId: string): {
    selectedEnv: string;
    setSelectedEnv: (env: string) => void;
} {
    const selectedEnv = useSelectedProjectEnv(projectId);
    const setSelectedEnv = useProjectEnvFilterStore(state => state.setSelectedEnv);

    return {
        selectedEnv,
        setSelectedEnv: env => {
            setSelectedEnv(projectId, env);
        },
    };
}

export function getProjectEnvFilterParam(selectedEnv: string): string | undefined {
    if (selectedEnv === PROJECT_ENV_FILTER_ALL) {
        return undefined;
    }

    return selectedEnv;
}
