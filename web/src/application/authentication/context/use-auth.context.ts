import { type Draft, produce } from "immer";
import { z } from "zod";
import { create } from "zustand";

const KEY = "2fa-authentication";

const Schema = z.object({
    email: z.string().trim().nonempty(),
    mfaToken: z.string().trim().nonempty(),
    required2FA: z.literal(true),
});

const MfaSetupRequiredSchema = z.object({
    mfaSetupRequired: z.literal(true),
});

class JSONStorage {
    persisted(defaultData: State["data"]): State["data"] {
        const value = localStorage.getItem(KEY);

        if (!value) {
            return defaultData;
        }

        try {
            const json = JSON.parse(value);

            // Try to parse as required2FA schema
            const parsed2FA = Schema.safeParse(json);
            if (parsed2FA.success) {
                return {
                    required2FA: parsed2FA.data.required2FA,
                    email: parsed2FA.data.email,
                    mfaToken: parsed2FA.data.mfaToken,
                };
            }

            // Try to parse as mfaSetupRequired schema
            const parsedMfaSetup = MfaSetupRequiredSchema.safeParse(json);
            if (parsedMfaSetup.success) {
                return {
                    mfaSetupRequired: parsedMfaSetup.data.mfaSetupRequired,
                };
            }

            // If neither matches, remove invalid data
            localStorage.removeItem(KEY);
        } catch {
            localStorage.removeItem(KEY);
        }

        return defaultData;
    }

    update(data: State["data"]): void {
        localStorage.setItem(KEY, JSON.stringify(data));
    }

    clear(): void {
        localStorage.removeItem(KEY);
    }
}

interface State {
    data:
        | {
              required2FA: true;
              email: string;
              mfaToken: string;
          }
        | {
              mfaSetupRequired: true;
          }
        | {
              required2FA: false;
              mfaSetupRequired: false;
          };
}

interface Actions {
    enable2FA: (params: { email: string; mfaToken: string }) => void;
    enableMfaSetup: () => void;
    clear: () => void;
}

export const useAuthContext = create<State & Actions>()(set => {
    const storage = new JSONStorage();

    return {
        data: storage.persisted({
            required2FA: false,
            mfaSetupRequired: false,
        }),

        enable2FA: params => {
            set(
                produce((draft: Draft<State>) => {
                    draft.data = {
                        required2FA: true,
                        email: params.email,
                        mfaToken: params.mfaToken,
                    };

                    storage.update(draft.data);
                }),
            );
        },

        enableMfaSetup: () => {
            set(
                produce((draft: Draft<State>) => {
                    draft.data = {
                        mfaSetupRequired: true,
                    };

                    storage.update(draft.data);
                }),
            );
        },

        clear: () => {
            set(
                produce((draft: Draft<State>) => {
                    draft.data = {
                        required2FA: false,
                        mfaSetupRequired: false,
                    };

                    storage.clear();
                }),
            );
        },
    };
});
