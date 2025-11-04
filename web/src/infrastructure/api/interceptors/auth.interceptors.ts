import { AxiosError, type AxiosInstance } from "axios";

export function initAuthInterceptors(client: AxiosInstance): void {
    /**
     * Auth request interceptor
     */
    client.interceptors.request.use(config => {
        const token = localStorage.getItem("token");

        if (token !== null) {
            config.headers.Authorization = `Bearer ${token}`;
        }

        return config;
    });

    /**
     * Auth response interceptor
     */
    client.interceptors.response.use(
        config => config,
        error => {
            // TODO refresh jwt token

            if (error instanceof AxiosError) {
                // console.log(error.response);
            }

            throw error;
        },
    );
}
