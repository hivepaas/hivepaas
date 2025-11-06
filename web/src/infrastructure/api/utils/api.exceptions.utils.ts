import { type AxiosError } from "axios";

import { CancelException } from "@infrastructure/exceptions/cancel";
import { Http401Exception, Http403Exception, Http404Exception, HttpException } from "@infrastructure/exceptions/http";

import { parseApiError } from "./api.data.utils";

/**
 * Check if the error is a cancel exception
 */
export function isCancelException(error: Error): error is CancelException {
    return error instanceof CancelException;
}

/**
 * Check if the error is a token expired error
 */
export function isTokenExpiredException(error: AxiosError): boolean {
    const err = parseApiError(error);

    return err instanceof Http401Exception && err.code === "ERR_SESSION_JWT_EXPIRED";
}

/**
 * Check if the error is a invite token invalid error
 */
export function isInviteTokenInvalidException(error: Error): boolean {
    return error instanceof Http403Exception && error.code === "ERR_USER_INVITE_TOKEN_INVALID";
}

/**
 * Check if the error is HttpException
 */
export function isHttpException(error: Error): error is HttpException {
    return error instanceof HttpException;
}

/**
 * Check if the error is a Http404Exception
 */
export function isHttp404Exception(error: Error): error is Http404Exception {
    return error instanceof Http404Exception;
}

/**
 * Check if the error is a unauthorized exception
 */
export function isUnauthorizedException(error: Error): boolean {
    return error instanceof Http401Exception && error.code === "ERR_UNAUTHORIZED";
}

/**
 * Check if the error is a session invalid error
 */
export function isSessionInvalidException(error: Error): boolean {
    return error instanceof Http401Exception && error.code === "ERR_SESSION_JWT_INVALID";
}

/**
 * Check if the error is a user unavailable error
 */
export function isUserUnavailableException(error: Error): boolean {
    return (
        error instanceof Http403Exception && ["ERR_USER_UNAVAILABLE", "ERR_USER_NOT_IN_WORKSPACE"].includes(error.code)
    );
}

/**
 * Check if the error is a too many login attempts error
 */
export function isToManyLoginAttemptsException(error: Error): boolean {
    return (
        error instanceof Http403Exception &&
        ["ERR_TOO_MANY_LOGIN_FAILURES", "ERR_TOO_MANY_PASSCODE_ATTEMPTS"].includes(error.code)
    );
}
