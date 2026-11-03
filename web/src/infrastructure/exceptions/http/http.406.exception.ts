import { type ProblemApiResponse } from "@infrastructure/api";

import { HttpException } from "./http.exception";

/**
 * Not Acceptable
 */
export class Http406Exception extends HttpException {
    constructor(problem: ProblemApiResponse) {
        super({
            status: 406,
            problem,
        });
    }
}
