interface ConstructorParams {
    type: string;
    title: string;
    status: number;
    detail: string;
}

export class ProblemApiResponse {
    constructor(params: ConstructorParams) {
        this.type = params.type;
        this.title = params.title;
        this.status = params.status;
        this.detail = params.detail;
    }

    type: string;
    title: string;
    status: number;
    detail: string;
}
