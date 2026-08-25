export class UnexpectedApiResponseException extends Error {
    public constructor() {
        super("Unexpected error happen");
    }
}
