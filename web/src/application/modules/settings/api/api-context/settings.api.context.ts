import { createContext } from "react";

import { NotificationsApi, NotificationsApiValidator } from "../services/notifications-services";

function createApi() {
    /**
     * Settings
     */
    const notificationsValidator = new NotificationsApiValidator();

    return {
        settings: {
            notifications: new NotificationsApi(notificationsValidator),
        },
    };
}

export const SettingsApiContext = createContext({
    api: createApi(),
});
