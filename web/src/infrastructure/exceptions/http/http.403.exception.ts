import { type ProblemApiResponse } from "@infrastructure/api";

import { HttpException } from "./http.exception";

/**
 * Forbidden
 */
export class Http403Exception extends HttpException {
    constructor(problem: ProblemApiResponse) {
        super({
            status: 403,
            problem,
        });
    }
}
