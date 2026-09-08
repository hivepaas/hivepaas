import { type ProblemApiResponse } from "@infrastructure/api";

interface ConstructorParams {
    status: number;
    problem: ProblemApiResponse;
    /**
     * The URL the failed request was sent to.
     *
     * Kept because some refusals are about the resource rather than the request,
     * and the handler that reacts to them is nowhere near the call site - it needs
     * a way back to the thing that was refused. See ERR_SETTING_IN_USE, where it
     * is what locates the list of blockers.
     */
    requestUrl?: string;
}

/**
 * Base HTTP exception
 */
export class HttpException extends Error {
    constructor({ status, problem, requestUrl }: ConstructorParams) {
        super(problem.detail);

        this.status = status;
        this.code = problem.code;
        this.problem = problem;
        this.requestUrl = requestUrl;
    }

    readonly status: number;
    readonly code: string;
    readonly requestUrl?: string;

    readonly problem: ProblemApiResponse;
}
