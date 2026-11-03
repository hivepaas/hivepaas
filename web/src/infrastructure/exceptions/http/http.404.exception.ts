import { type ProblemApiResponse } from "@infrastructure/api";

import { HttpException } from "./http.exception";

/**
 * Not found
 */
export class Http404Exception extends HttpException {
    constructor(problem: ProblemApiResponse) {
        super({
            status: 404,
            problem,
        });
    }
}
