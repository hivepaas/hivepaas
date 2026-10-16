interface ConstructorParams {
    type?: string;
    title: string;
    status: number;
    code: string;
    detail: string;
    displayLevel?: "high" | "medium" | "low";
    params?: Record<string, unknown> | null;
}

export class ProblemApiResponse {
    constructor(params: ConstructorParams) {
        this.type = params.type ?? "";
        this.title = params.title;
        this.status = params.status;
        this.code = params.code;
        this.detail = params.detail;
        this.displayLevel = params.displayLevel;
        this.params = params.params ?? {};
    }

    readonly type?: string;
    readonly title: string;
    readonly status: number;
    readonly code: string;
    readonly detail: string;
    readonly displayLevel?: "high" | "medium" | "low";

    /**
     * Values the client may act on, beside the message they appear in: the limit
     * a refused creation reached, for one - ERR_PLAN_LIMIT_REACHED sends Limit,
     * Max and Plan. Which a problem has is part of what its code means, so read
     * them by the code, and check their shape: they are as the server sent them.
     * Empty when there are none, which is most problems.
     */
    readonly params: Readonly<Record<string, unknown>>;
}
