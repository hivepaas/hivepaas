interface ConstructorParams {
    type?: string;
    title: string;
    status: number;
    code: string;
    detail: string;
    displayLevel?: "high" | "medium" | "low";
}

export class ProblemApiResponse {
    constructor(params: ConstructorParams) {
        this.type = params.type ?? "";
        this.title = params.title;
        this.status = params.status;
        this.code = params.code;
        this.detail = params.detail;
        this.displayLevel = params.displayLevel;
    }

    readonly type?: string;
    readonly title: string;
    readonly status: number;
    readonly code: string;
    readonly detail: string;
    readonly displayLevel?: "high" | "medium" | "low";
}
