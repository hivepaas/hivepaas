import { AxiosError, type AxiosResponse } from "axios";
import { z } from "zod";

import { ProblemApiResponse, ValidationProblemApiResponse } from "@infrastructure/api/types";

import {
    Http400Exception,
    Http401Exception,
    Http403Exception,
    Http404Exception,
    Http409Exception,
    Http422Exception,
    Http500Exception,
    HttpException,
} from "@infrastructure/exceptions/http";

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
        if (import.meta.env.DEV) {
            console.warn(response.data, parsed.error.flatten());
        }

        throw new Error("API response validation error.");
    }

    return parsed.data;
}

const ProblemApiResponseSchema = z.object({
    type: z.string(),
    title: z.string(),
    status: z.number(),
    detail: z.string(),
});

const ValidationProblemApiResponseSchema = z
    .object({
        errors: z.array(
            z.object({
                name: z.string(),
                code: z.string(),
                message: z.string(),
            }),
        ),
    })
    .merge(ProblemApiResponseSchema);

/**
 * Handle api error
 */
export function handleApiError(error: unknown): Error {
    if (error instanceof AxiosError) {
        const data: unknown = error.response?.data;

        const parsedValidationProblem = ValidationProblemApiResponseSchema.safeParse(data);

        if (parsedValidationProblem.success) {
            const problem = new ValidationProblemApiResponse(parsedValidationProblem.data);

            return new Http400Exception(problem.detail);
        }

        const parsedProblem = ProblemApiResponseSchema.safeParse(data);

        if (parsedProblem.success) {
            const problem = new ProblemApiResponse(parsedProblem.data);

            if (problem.status === 400) {
                return new Http400Exception(problem.detail);
            }

            if (problem.status === 401) {
                return new Http401Exception(problem.detail);
            }

            if (problem.status === 403) {
                return new Http403Exception(problem.detail);
            }

            if (problem.status === 404) {
                return new Http404Exception(problem.detail);
            }

            if (problem.status === 409) {
                return new Http409Exception(problem.detail);
            }

            if (problem.status === 422) {
                return new Http422Exception(problem.detail);
            }

            if (problem.status === 500) {
                return new Http500Exception(problem.detail);
            }

            return new HttpException({
                message: problem.detail,
                status: problem.status,
            });
        }
    }

    if (error instanceof Error) {
        return error;
    }

    return new Error("Unexpected error.");
}
