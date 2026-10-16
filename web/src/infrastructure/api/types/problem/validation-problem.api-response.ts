import { ProblemApiResponse } from "@infrastructure/api";

interface ValidationError {
    path: string;
    code: string;
    message: string;
}

interface ConstructorParams {
    type?: string;
    title: string;
    status: number;
    code: string;
    detail: string;
    errors: ValidationError[];
    displayLevel?: "high" | "medium" | "low";
    params?: Record<string, unknown> | null;
}

export class ValidationProblemApiResponse extends ProblemApiResponse {
    constructor(params: ConstructorParams) {
        super({
            type: params.type ?? "",
            title: params.title,
            status: params.status,
            code: params.code,
            detail: params.detail,
            displayLevel: params.displayLevel,
            params: params.params,
        });

        this.#errors = params.errors;
    }

    #errors: ValidationError[];

    get errors(): ValidationError[] {
        return this.#errors;
    }

    set errors(errors: ValidationError[]) {
        this.#errors = errors;
    }
}
