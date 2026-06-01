import { type Draft, produce } from "immer";
import { create } from "zustand";

import { type Profile } from "@application/shared/entities";
import { EUserRole } from "@application/shared/enums";
import { useModulePermissionsStore, useProjectPermissionsStore } from "@application/shared/permissions/store";

import { useAuthContext } from "@application/authentication/context";

interface State {
    profile: Profile | null;
}

interface Actions {
    setProfile: (profile: Profile) => void;
    updateProfile: (profile: Partial<Profile>) => void;
    clearProfile: () => void;
}

export const useProfileContext = create<State & Actions>()(set => ({
    profile: null,

    setProfile: profile => {
        set(
            produce((draft: Draft<State>) => {
                draft.profile = profile;
            }),
        );
    },

    updateProfile: profile => {
        set(
            produce((draft: Draft<State>) => {
                if (draft.profile === null) {
                    return;
                }

                if (profile.username) {
                    draft.profile.username = profile.username;
                }

                if (profile.email) {
                    draft.profile.email = profile.email;
                }

                if (profile.photo) {
                    draft.profile.photo = profile.photo;
                }

                if (profile.securityOption) {
                    draft.profile.securityOption = profile.securityOption;
                }

                if (profile.fullName) {
                    draft.profile.fullName = profile.fullName;
                }
            }),
        );
    },

    clearProfile: () => {
        set(
            produce((draft: Draft<State>) => {
                draft.profile = null;
            }),
        );
    },
}));

useProfileContext.subscribe((state, prevState) => {
    if (state.profile === prevState.profile) {
        return;
    }

    const auth = useAuthContext.getState();
    const modulePermissions = useModulePermissionsStore.getState();
    const projectPermissions = useProjectPermissionsStore.getState();

    if (state.profile !== null) {
        if (state.profile.role === EUserRole.Admin) {
            modulePermissions.setFullModules();
            projectPermissions.clearProjects();
        } else {
            modulePermissions.setModules(state.profile.modulePermissions);
            projectPermissions.setProjects(state.profile.projectPermissions);
        }

        auth.clear();

        return;
    }

    modulePermissions.clearModules();
    projectPermissions.clearProjects();
});
