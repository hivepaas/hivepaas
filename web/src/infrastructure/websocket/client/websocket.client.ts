import { EnvConfig } from "@config";

export type WebSocketReadyState = WebSocket["readyState"];

export type WebSocketQueryValue = string | number | boolean | null | undefined;

export type WebSocketQuery = Record<string, WebSocketQueryValue>;

export interface WebSocketHandlers {
    onOpen?: (event: Event, socket: WebSocket) => void;
    onMessage?: (message: string, event: MessageEvent<unknown>, socket: WebSocket) => void;
    onMessageError?: (error: unknown, socket: WebSocket) => void;
    onClose?: (event: CloseEvent, socket: WebSocket) => void;
    onError?: (event: Event, socket: WebSocket) => void;
    onReadyStateChange?: (readyState: WebSocketReadyState, socket: WebSocket) => void;
}

export interface WebSocketConnectOptions {
    signal?: AbortSignal;
    closeOnError?: boolean;
}

export interface WebSocketSubscription {
    readonly socket: WebSocket;
    getReadyState: () => WebSocketReadyState;
    close: (code?: number, reason?: string) => void;
}

export interface WebSocketClient {
    buildUrl: (path: string, query?: WebSocketQuery) => string;
    connect: (url: string, handlers?: WebSocketHandlers, options?: WebSocketConnectOptions) => WebSocketSubscription;
}

export function buildWebSocketUrl(path: string, query: WebSocketQuery = {}): string {
    const baseUrl = EnvConfig.API_URL.endsWith("/") ? EnvConfig.API_URL : `${EnvConfig.API_URL}/`;
    const url = new URL(path.replace(/^\/+/, ""), baseUrl);

    url.protocol = url.protocol === "https:" ? "wss:" : "ws:";

    Object.entries(query).forEach(([key, value]) => {
        if (value === null || value === undefined) {
            return;
        }

        url.searchParams.set(key, String(value));
    });

    return url.toString();
}

export async function readWebSocketMessageData(data: unknown): Promise<string> {
    if (typeof data === "string") {
        return data;
    }

    if (data instanceof Blob) {
        return data.text();
    }

    if (data instanceof ArrayBuffer) {
        return new TextDecoder().decode(data);
    }

    return "";
}

export function createWebSocketClient(): WebSocketClient {
    return {
        buildUrl: buildWebSocketUrl,
        connect,
    };
}

function connect(
    url: string,
    handlers: WebSocketHandlers = {},
    options: WebSocketConnectOptions = {},
): WebSocketSubscription {
    const { signal, closeOnError = true } = options;
    const socket = new WebSocket(url);
    let isCleanedUp = false;

    function emitReadyState(): void {
        handlers.onReadyStateChange?.(socket.readyState, socket);
    }

    function cleanup(): void {
        if (isCleanedUp) {
            return;
        }

        isCleanedUp = true;
        signal?.removeEventListener("abort", handleAbort);
        socket.removeEventListener("open", handleOpen);
        socket.removeEventListener("message", handleMessage);
        socket.removeEventListener("close", handleClose);
        socket.removeEventListener("error", handleError);
    }

    function close(code?: number, reason?: string): void {
        signal?.removeEventListener("abort", handleAbort);

        if (socket.readyState === WebSocket.CLOSING || socket.readyState === WebSocket.CLOSED) {
            cleanup();
            return;
        }

        socket.close(code, reason);
    }

    function handleAbort(): void {
        close();
    }

    function handleOpen(event: Event): void {
        emitReadyState();
        handlers.onOpen?.(event, socket);
    }

    function handleMessage(event: MessageEvent<unknown>): void {
        void readWebSocketMessageData(event.data)
            .then(message => {
                handlers.onMessage?.(message, event, socket);
            })
            .catch((error: unknown) => {
                handlers.onMessageError?.(error, socket);
            });
    }

    function handleClose(event: CloseEvent): void {
        emitReadyState();
        handlers.onClose?.(event, socket);
        cleanup();
    }

    function handleError(event: Event): void {
        handlers.onError?.(event, socket);

        if (closeOnError && socket.readyState !== WebSocket.CLOSING && socket.readyState !== WebSocket.CLOSED) {
            socket.close();
        }

        emitReadyState();
    }

    socket.addEventListener("open", handleOpen);
    socket.addEventListener("message", handleMessage);
    socket.addEventListener("close", handleClose);
    socket.addEventListener("error", handleError);

    if (signal?.aborted) {
        close();
    } else {
        signal?.addEventListener("abort", handleAbort, { once: true });
    }

    emitReadyState();

    return {
        socket,
        getReadyState: () => socket.readyState,
        close,
    };
}
