export const QK = {
    /**
     * Session
     */
    "session.get-profile": "session.get-profile",

    /*
     * Profile API Keys
     */
    "profile-api-keys.$.find-many-paginated": "profile-api-keys.$.find-many-paginated",

    /*
     * License: one key for System › License, the banners, the plan-limit dialog
     * and the OAuth page, so one invalidation updates them all.
     */
    "license.find-one": "license.find-one",
} as const;
