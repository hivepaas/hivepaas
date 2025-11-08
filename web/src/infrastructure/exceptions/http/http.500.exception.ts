import { type ProblemApiResponse } from "@infrastructure/api";

import { HttpException } from "./http.exception";

/**
 * Internal server error
 */
export class Http500Exception extends HttpException {
    constructor(problem: ProblemApiResponse) {
        super({
            status: 500,
            problem,
        });
    }
}
