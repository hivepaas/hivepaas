export class UnexpectedApiResponseException extends Error {
    public constructor() {
        super("Unexpected API response.");
    }
}
