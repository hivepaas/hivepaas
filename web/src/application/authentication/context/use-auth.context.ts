import { type Draft, produce } from "immer";
import { z } from "zod";
import { create } from "zustand";

const KEY = "2fa-authentication";

const Schema = z.object({
    email: z.string().trim().nonempty(),
    mfaToken: z.string().trim().nonempty(),
    required2FA: z.literal(true),
});

class JSONStorage {
    persisted(defaultData: State["data"]): State["data"] {
        const value = localStorage.getItem(KEY);

        if (!value) {
            return defaultData;
        }

        try {
            const json = JSON.parse(value);

            const parsed = Schema.parse(json);

            return {
                required2FA: parsed.required2FA,
                email: parsed.email,
                mfaToken: parsed.mfaToken,
            };
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
              required2FA: false;
          };
}

interface Actions {
    enable2FA: (params: { email: string; mfaToken: string }) => void;
    clear: () => void;
}

export const useAuthContext = create<State & Actions>()(set => {
    const storage = new JSONStorage();

    return {
        data: storage.persisted({
            required2FA: false,
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

        clear: () => {
            set(
                produce((draft: Draft<State>) => {
                    draft.data = {
                        required2FA: false,
                    };

                    storage.clear();
                }),
            );
        },
    };
});
