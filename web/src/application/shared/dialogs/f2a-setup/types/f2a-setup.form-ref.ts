import { type ValidationException } from "@infrastructure/exceptions/validation";

export interface F2aSetupFormRef {
    onError: (error: ValidationException) => void;
}
