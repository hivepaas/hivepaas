import { type ProblemApiResponse } from "@infrastructure/api";

interface ConstructorParams {
    status: number;
    problem: ProblemApiResponse;
}

/**
 * Base HTTP exception
 */
export class HttpException extends Error {
    constructor({ status, problem }: ConstructorParams) {
        super(problem.detail);

        this.status = status;
        this.code = problem.code;
        this.problem = problem;
    }

    readonly status: number;
    readonly code: string;

    readonly problem: ProblemApiResponse;
}
