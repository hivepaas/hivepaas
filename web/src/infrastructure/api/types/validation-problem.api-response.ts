import { ProblemApiResponse } from "./problem.api-response";

interface ValidationError {
    name: string;
    code: string;
    message: string;
}

interface ConstructorParams {
    type: string;
    title: string;
    status: number;
    detail: string;
    errors: ValidationError[];
}

export class ValidationProblemApiResponse extends ProblemApiResponse {
    constructor(params: ConstructorParams) {
        super({
            type: params.type,
            title: params.title,
            status: params.status,
            detail: params.detail,
        });

        this.errors = params.errors;
    }

    errors: ValidationError[];
}
