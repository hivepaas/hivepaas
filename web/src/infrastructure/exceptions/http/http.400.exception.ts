import { type ProblemApiResponse, ValidationProblemApiResponse } from "@infrastructure/api";

import { HttpException } from "./http.exception";

interface ValidationError {
    path: string;
    code: string;
    message: string;
}

/**
 * Bad request
 */
export class Http400Exception extends HttpException {
    constructor(problem: ProblemApiResponse) {
        super({
            status: 400,
            problem,
        });
    }

    get hasErrors(): boolean {
        if (this.problem instanceof ValidationProblemApiResponse) {
            return this.problem.errors.length > 0;
        }

        return false;
    }

    get errors(): ValidationError[] {
        if (this.problem instanceof ValidationProblemApiResponse) {
            return this.problem.errors;
        }

        return [];
    }

    set errors(errors: ValidationError[]) {
        if (this.problem instanceof ValidationProblemApiResponse) {
            this.problem.errors = errors;
        }
    }
}
