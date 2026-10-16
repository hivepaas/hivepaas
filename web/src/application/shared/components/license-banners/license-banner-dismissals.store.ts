import { create } from "zustand";

const STORAGE_KEY = "hivepaas-license-banners-dismissed";

/** Until when each dismissed banner stays hidden, in milliseconds since the epoch, by the banner's id. */
type LicenseBannerDismissals = Record<string, number>;

/** The dismissals still holding at now. */
function holding(dismissals: LicenseBannerDismissals, now: number): LicenseBannerDismissals {
    return Object.fromEntries(Object.entries(dismissals).filter(([, until]) => until > now));
}

/**
 * Dismissals are kept in this browser only: dismissing a banner is a person's
 * say about their own screen, not about the installation. Storage that cannot be
 * used - a private window, storage turned off - is quietly done without: a
 * dismissal then lasts until the page is loaded again.
 */
class LicenseBannerDismissalsStorage {
    /** Those still holding at now. The rest are dropped from storage, so it never grows. */
    persisted(now: number): LicenseBannerDismissals {
        try {
            const value = window.localStorage.getItem(STORAGE_KEY);

            if (!value) {
                return {};
            }

            const json = JSON.parse(value) as unknown;

            if (!json || typeof json !== "object" || Array.isArray(json)) {
                this.update({});
                return {};
            }

            const entries = Object.entries(json);
            const dismissals = entries.reduce<LicenseBannerDismissals>((acc, [id, until]) => {
                if (typeof until !== "number" || !Number.isFinite(until) || until <= now) {
                    return acc;
                }

                return {
                    ...acc,
                    [id]: until,
                };
            }, {});

            if (Object.keys(dismissals).length !== entries.length) {
                this.update(dismissals);
            }

            return dismissals;
        } catch {
            // Not JSON, or no storage to read: start afresh.
            this.update({});
            return {};
        }
    }

    update(dismissals: LicenseBannerDismissals): void {
        try {
            if (Object.keys(dismissals).length === 0) {
                window.localStorage.removeItem(STORAGE_KEY);
                return;
            }

            window.localStorage.setItem(STORAGE_KEY, JSON.stringify(dismissals));
        } catch {
            // Nowhere to keep it: see the class.
        }
    }
}

interface State {
    dismissed: LicenseBannerDismissals;
}

interface Actions {
    /** Hides the banner with this id for snoozeMs, from now. */
    dismiss: (id: string, snoozeMs: number) => void;
}

const storage = new LicenseBannerDismissalsStorage();

export const useLicenseBannerDismissalsStore = create<State & Actions>()(set => ({
    dismissed: storage.persisted(Date.now()),

    dismiss: (id, snoozeMs) => {
        set(state => {
            const now = Date.now();
            const dismissed = {
                ...holding(state.dismissed, now),
                [id]: now + snoozeMs,
            };

            storage.update(dismissed);

            return { dismissed };
        });
    },
}));

/** Whether the banner with this id was dismissed, and the dismissal still holds at now. */
export function isLicenseBannerDismissed(
    dismissed: Readonly<LicenseBannerDismissals>,
    id: string,
    now: number,
): boolean {
    return (dismissed[id] ?? 0) > now;
}
