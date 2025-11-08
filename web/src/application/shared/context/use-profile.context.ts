import { type Draft, produce } from "immer";
import { create } from "zustand";

import { type Profile } from "@application/shared/entities";

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

                if (profile.email !== undefined) {
                    draft.profile.email = profile.email;
                }

                if (profile.photo !== undefined) {
                    draft.profile.photo = profile.photo;
                }

                if (profile.securityOption !== undefined) {
                    draft.profile.securityOption = profile.securityOption;
                }

                if (profile.fullName !== undefined) {
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

    if (state.profile !== null) {
        auth.clear();
    }
});
