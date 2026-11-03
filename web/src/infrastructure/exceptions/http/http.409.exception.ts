import { type ProblemApiResponse } from "@infrastructure/api";

import { HttpException } from "./http.exception";

/**
 * Conflict
 */
export class Http409Exception extends HttpException {
    constructor(problem: ProblemApiResponse) {
        super({
            status: 409,
            problem,
        });
    }
}
