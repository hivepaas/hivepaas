import { type AxiosInstance, type InternalAxiosRequestConfig, isAxiosError } from "axios";
import { match } from "oxide.ts";

import {
    isSessionInvalidException,
    isSessionOverException,
    isTokenExpired,
    isTokenExpiredException,
    refreshToken,
    session,
} from "@infrastructure/api";
import { RefreshQueue } from "@infrastructure/api/queues";

function setAuthorizationHeader(request: InternalAxiosRequestConfig, token: string) {
    request.headers.Authorization = `Bearer ${token}`; // eslint-disable-line no-param-reassign
}

function isAuthPath(pathname: string) {
    const normalizedPathname = pathname.replace(/\/+$/, "");

    return normalizedPathname === "/auth" || normalizedPathname.startsWith("/auth/");
}

function redirectToSignIn() {
    session.removeToken();

    if (isAuthPath(window.location.pathname)) {
        return;
    }

    const next = `${window.location.pathname}${window.location.search}${window.location.hash}`;

    window.location.href = `/auth/sign-in/?next=${encodeURIComponent(next)}`;
}

const attempts = new WeakMap<object, boolean>();
const refreshQueue = new RefreshQueue();

export function initAuthInterceptors(client: AxiosInstance): void {
    /**
     * Auth request interceptor
     */
    client.interceptors.request.use(async request => {
        const token = session.getToken();

        /*
         * Skip if access token is not set
         */
        if (token === null) {
            return request;
        }

        /*
         * Skip if access token is not expired
         */
        if (!isTokenExpired(token)) {
            setAuthorizationHeader(request, token);

            return request;
        }

        /*
         * Add the request to the queue if the token is refreshing now
         */
        if (refreshQueue.isOpen) {
            return new Promise(resolve => {
                refreshQueue.add({
                    onSuccess: newToken => {
                        setAuthorizationHeader(request, newToken);

                        resolve(request);
                    },
                    onFailed: () => {
                        resolve(request);
                    },
                });
            });
        }

        refreshQueue.open();

        /*
         * Try to refresh the access token before sending the request
         */
        const result = await refreshToken();

        return match(result, {
            Ok: newToken => {
                refreshQueue.success(newToken);

                session.setToken(newToken);

                setAuthorizationHeader(request, newToken);

                return request;
            },
            Err: error => {
                refreshQueue.failed();

                /*
                 * Only when the server says the session is finished. A refresh
                 * that failed because nothing answered it says nothing about the
                 * session - the cookie holding it is still good - so the request
                 * goes out and fails on its own terms, and the next one tries
                 * refreshing again. Signing the person out here would throw away
                 * a working session, and their place in the app, over a blip.
                 */
                if (isSessionOverException(error)) {
                    redirectToSignIn();
                }

                return request;
            },
        });
    });

    /**
     * Auth response interceptor
     */
    client.interceptors.response.use(
        response => response,
        async (error: Error) => {
            /*
             * Skip if the error is not an Axios error or the request is not set
             */
            if (!isAxiosError(error) || error.config === undefined) {
                return Promise.reject(error);
            }

            if (isSessionInvalidException(error)) {
                redirectToSignIn();

                return Promise.reject(error);
            }

            /**
             * Skip if the error is not a "token expired error"
             */
            if (!isTokenExpiredException(error)) {
                return Promise.reject(error);
            }

            const request = error.config;

            /*
             * Skip if request has already been attempted to prevent infinite loop
             */
            if (attempts.has(request)) {
                return Promise.reject(error);
            }

            attempts.set(request, true);

            /**
             * Add the request to the queue if the token is refreshing now
             */
            if (refreshQueue.isOpen) {
                return new Promise((resolve, reject) => {
                    refreshQueue.add({
                        onSuccess: newToken => {
                            setAuthorizationHeader(request, newToken);

                            resolve(client(request));
                        },
                        onFailed: () => {
                            reject(error);
                        },
                    });
                });
            }

            refreshQueue.open();

            const result = await refreshToken();

            /*
             * Try to refresh the access token and resend the request
             */
            return match(result, {
                Ok: newToken => {
                    refreshQueue.success(newToken);

                    session.setToken(newToken);

                    setAuthorizationHeader(request, newToken);

                    return client(request);
                },
                Err: refreshError => {
                    refreshQueue.failed();

                    // Same rule as the request side: the session ends when the
                    // server says so, not when the network drops.
                    if (isSessionOverException(refreshError)) {
                        redirectToSignIn();
                    }

                    return Promise.reject(error);
                },
            });
        },
    );
}
