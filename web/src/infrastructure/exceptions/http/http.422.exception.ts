import { type ProblemApiResponse } from "@infrastructure/api";

import { HttpException } from "./http.exception";

/**
 * Unprocessable Entity
 */
export class Http422Exception extends HttpException {
    constructor(problem: ProblemApiResponse) {
        super({
            status: 422,
            problem,
        });
    }
}
