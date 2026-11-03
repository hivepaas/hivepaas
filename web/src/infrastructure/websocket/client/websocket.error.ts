/**
 * Normalises an unknown thrown value from a WebSocket operation into an `Error`.
 *
 * `new WebSocket(url)` can throw synchronously for malformed URLs.
 * Every WS service that wraps `client.connect` in a try/catch should
 * call this helper instead of defining its own local conversion.
 */
export function toWebSocketError(error: unknown, fallbackMessage: string = "WebSocket operation failed."): Error {
    if (error instanceof Error) {
        return error;
    }

    return new Error(fallbackMessage);
}
