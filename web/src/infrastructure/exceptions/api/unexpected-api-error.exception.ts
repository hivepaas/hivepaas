export class UnexpectedApiErrorException extends Error {
    public constructor() {
        super("Unexpected API error response.");
    }
}
