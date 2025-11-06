import { Http400Exception, type HttpException } from "@infrastructure/exceptions/http";

interface ValidationError {
    path: string;
    code: string;
    message: string;
}

/**
 * Validation exception
 */
export class ValidationException extends Error {
    private constructor(errors: ValidationError[]) {
        super("Validation failed");

        this.errors = errors;
    }

    readonly errors: ValidationError[];

    public static fromHttp(error: HttpException): ValidationException {
        if (error instanceof Http400Exception && error.hasErrors) {
            return new ValidationException(error.errors);
        }

        return ValidationException.empty;
    }

    public static get empty(): ValidationException {
        return new ValidationException([]);
    }
}
