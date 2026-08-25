export class UnexpectedApiErrorException extends Error {
    public constructor() {
        super("Unexpected error happen");
    }
}
