import { type Draft, produce } from "immer";
import { create } from "zustand";

import { type Profile } from "@application/shared/entities";

interface State {
  profile: Profile | null;
  token: string | null;
}

interface Actions {
  setProfile: (profile: Profile) => void;
  setToken: (token: string) => void;
  reset: () => void;
}

export const useProfileContext = create<State & Actions>()((set) => ({
  profile: null,
  token: localStorage.getItem("token"),

  setProfile: (profile) => {
    set(
      produce((draft: Draft<State>) => {
        draft.profile = profile;
      })
    );
  },

  setToken: (token) => {
    set(
      produce((draft: Draft<State>) => {
        draft.token = token;

        localStorage.setItem("token", token);
      })
    );
  },

  reset: () => {
    set(
      produce((draft: Draft<State>) => {
        draft.profile = null;
        draft.token = null;

        localStorage.removeItem("token");
      })
    );
  },
}));

useProfileContext.subscribe((state, prevState) => {
  if (state.profile === prevState.profile) {
    return;
  }
});
