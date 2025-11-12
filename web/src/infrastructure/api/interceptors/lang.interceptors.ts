import { type AxiosInstance } from "axios";

export function initLangInterceptors(client: AxiosInstance): void {
    /**
     * Lang request interceptor
     */
    client.interceptors.request.use(config => {
        const lang = localStorage.getItem("lang");

        if (lang !== null) {
            config.headers["Accept-Language"] = lang; // eslint-disable-line no-param-reassign
        }

        return config;
    });
}
