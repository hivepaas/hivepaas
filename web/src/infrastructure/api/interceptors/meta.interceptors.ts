import { type AxiosInstance } from "axios";

import { notifyApiMeta } from "@infrastructure/api/notifications/api-meta.notifications";
import { extractMetaNotifications } from "@infrastructure/api/utils";

// Extend AxiosRequestConfig so individual calls can opt out of meta notifications.
declare module "axios" {
    interface InternalAxiosRequestConfig {
        skipMetaNotifications?: boolean;
    }
}

export function initMetaInterceptors(client: AxiosInstance): void {
    client.interceptors.response.use(response => {
        if (response.config.skipMetaNotifications) {
            return response;
        }

        try {
            const meta = extractMetaNotifications(response.data);

            if (meta) {
                notifyApiMeta(meta);
            }
        } catch {
            // Never reject a successful response due to a notification error.
        }

        return response;
    });
}
