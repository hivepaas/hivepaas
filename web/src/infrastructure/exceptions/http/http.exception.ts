interface ConstructorParams {
    message: string;
    status: number;
}

export class HttpException extends Error {
    constructor(params: ConstructorParams) {
        super(params.message);

        this.status = params.status;
    }

    readonly status: number;
}
