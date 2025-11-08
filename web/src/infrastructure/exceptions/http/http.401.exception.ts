import { type ProblemApiResponse } from "@infrastructure/api";

import { HttpException } from "./http.exception";

/**
 * Unauthorized
 */
export class Http401Exception extends HttpException {
    constructor(problem: ProblemApiResponse) {
        super({
            status: 401,
            problem,
        });
    }
}
