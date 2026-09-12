import { type AxiosError, isAxiosError } from "axios";

import { CancelException } from "@infrastructure/exceptions/cancel";
import {
    Http400Exception,
    Http401Exception,
    Http403Exception,
    Http404Exception,
    HttpException,
} from "@infrastructure/exceptions/http";
import { NetworkException } from "@infrastructure/exceptions/network";
import { TimeoutException } from "@infrastructure/exceptions/timeout";

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
 *
 * This matched ERR_USER_INVITE_TOKEN_INVALID on a 403, and neither half was
 * ever sent: POST /users/signup-begin answers an unusable token with
 * hperrors.ErrTokenInvalid, which carries the code ERR_TOKEN_INVALID and a 412,
 * so the branch this guards could not run and an expired invite showed the
 * generic error screen.
 *
 * 412 has no exception class of its own and lands on the HttpException
 * catch-all, so the status is not what identifies it. The code is - and while
 * ERR_TOKEN_INVALID is shared with other token checks, the only caller is the
 * sign-up route, where the invite token is the only token in play.
 */
export function isInviteTokenInvalidException(error: Error): boolean {
    return error instanceof HttpException && error.code === "ERR_TOKEN_INVALID";
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
    if (isAxiosError(error)) {
        const err = parseApiError(error);

        return err instanceof Http401Exception && err.code === "ERR_SESSION_JWT_INVALID";
    }

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

/**
 * Check if the error is a validation error
 */
export function isValidationException(error: Error): error is Http400Exception {
    return error instanceof Http400Exception && error.hasErrors;
}

/**
 * Check if the error is a high level exception
 */
export function isHighLevelException(error: Error): boolean {
    return isHttpException(error) && error.problem.displayLevel === "high";
}

/**
 * Check if the error is a feature disabled error
 */
export function isFeatureDisabledException(error: Error): boolean {
    return isHttpException(error) && error.code === "ERR_FEATURE_DISABLED";
}

/**
 * Check if the error says nothing about the request and may work on a retry
 *
 * No response at all, a timeout, or the server answering 5xx: the request never
 * got an answer about itself, so nothing has been decided and asking again is
 * reasonable. Anything the server answered with a 4xx is a decision, and
 * repeating it would only get the same decision back.
 */
export function isTransientException(error: Error): boolean {
    if (error instanceof NetworkException || error instanceof TimeoutException) {
        return true;
    }

    return error instanceof HttpException && error.status >= 500;
}

/**
 * Check if the error is the server saying the session is over
 *
 * Only a 4xx counts. Losing a session costs the person their place and whatever
 * they were typing, so it takes the server actually saying so - not a network
 * that dropped, not a gateway restarting, and not a response we could not parse.
 * Anything else leaves the session alone and lets the next request try again.
 */
export function isSessionOverException(error: Error): boolean {
    return error instanceof HttpException && error.status >= 400 && error.status < 500;
}
