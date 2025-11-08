import { AxiosError, type AxiosResponse, isAxiosError } from "axios";
import { type z } from "zod";

import { ProblemApiResponse, ValidationProblemApiResponse } from "@infrastructure/api/types";
import { ProblemApiResponseSchema, ValidationProblemApiResponseSchema } from "@infrastructure/api/validation";

import { UnexpectedApiErrorException, UnexpectedApiResponseException } from "@infrastructure/exceptions/api";
import { CancelException } from "@infrastructure/exceptions/cancel";
import {
    Http400Exception,
    Http401Exception,
    Http403Exception,
    Http404Exception,
    Http406Exception,
    Http409Exception,
    Http422Exception,
    Http500Exception,
    HttpException,
} from "@infrastructure/exceptions/http";
import { NetworkException } from "@infrastructure/exceptions/network";
import { RouteNotFoundException } from "@infrastructure/exceptions/not-found";
import { TimeoutException } from "@infrastructure/exceptions/timeout";

interface ParseApiResponseParams<T extends z.SomeZodObject> {
    response: AxiosResponse;
    schema: T;
}

/**
 * Parse api response
 * @throws Error
 */
export function parseApiResponse<T extends z.SomeZodObject>(params: ParseApiResponseParams<T>): z.output<T> {
    const { response, schema } = params;

    const parsed = schema.safeParse(response.data);

    if (!parsed.success) {
        if (import.meta.env["NODE_ENV"] !== "production") {
            console.warn(response.data, parsed.error.format());
        }

        throw new UnexpectedApiResponseException();
    }

    return parsed.data;
}

/**
 * Parse api error
 */
export function parseApiError(error: unknown): Error {
    /**
     * Unexpected API response
     * @see parseApiResponse
     */
    if (error instanceof UnexpectedApiResponseException) {
        return error;
    }

    /**
     * Should never happen!
     */
    if (!isAxiosError(error)) {
        return new Error("Invalid API Client error.");
    }

    /**
     * Network error
     */
    if (error.code === AxiosError.ERR_NETWORK) {
        return new NetworkException(error.message);
    }

    /**
     * Timeout error
     */
    if (error.code === AxiosError.ETIMEDOUT) {
        return new TimeoutException(error.message);
    }

    /**
     * Cancel error
     */
    if (error.code === AxiosError.ERR_CANCELED) {
        return new CancelException(error.message);
    }

    const data: unknown = error.response?.data;

    /**
     * Route not found error
     */
    if (error.status === 404 && typeof data !== "object") {
        return new RouteNotFoundException(error.config?.url ?? "");
    }

    /**
     * Validation problem
     */
    const parsedValidationProblem = ValidationProblemApiResponseSchema.safeParse(data);

    if (parsedValidationProblem.success) {
        const problem = new ValidationProblemApiResponse(parsedValidationProblem.data);

        return new Http400Exception(problem);
    }

    /**
     * Problem
     */
    const parsedProblem = ProblemApiResponseSchema.safeParse(data);

    if (parsedProblem.success) {
        const problem = new ProblemApiResponse(parsedProblem.data);

        if (problem.status === 400) {
            return new Http400Exception(problem);
        }

        if (problem.status === 401) {
            return new Http401Exception(problem);
        }

        if (problem.status === 403) {
            return new Http403Exception(problem);
        }

        if (problem.status === 404) {
            return new Http404Exception(problem);
        }

        if (problem.status === 406) {
            return new Http406Exception(problem);
        }

        if (problem.status === 409) {
            return new Http409Exception(problem);
        }

        if (problem.status === 422) {
            return new Http422Exception(problem);
        }

        if (problem.status === 500) {
            return new Http500Exception(problem);
        }

        return new HttpException({
            status: problem.status,
            problem,
        });
    }

    /**
     * Unexpected API error response
     */
    return new UnexpectedApiErrorException();
}
