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
    protocols?: string | string[];
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
    const { signal, closeOnError = true, protocols } = options;
    const socket = new WebSocket(url, protocols);
    socket.binaryType = "arraybuffer";

    let isCleanedUp = false;
    let pendingMessageDispatch: Promise<void> = Promise.resolve();

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
        pendingMessageDispatch = pendingMessageDispatch.then(
            () => dispatchMessage(event),
            () => dispatchMessage(event),
        );
    }

    function handleClose(event: CloseEvent): void {
        void handleCloseAfterMessages(event);
    }

    function handleError(event: Event): void {
        handlers.onError?.(event, socket);

        if (closeOnError && socket.readyState !== WebSocket.CLOSING && socket.readyState !== WebSocket.CLOSED) {
            socket.close();
        }

        emitReadyState();
    }

    async function dispatchMessage(event: MessageEvent<unknown>): Promise<void> {
        try {
            const message = await readWebSocketMessageData(event.data);
            handlers.onMessage?.(message, event, socket);
        } catch (error: unknown) {
            reportMessageError(error);
        }
    }

    async function handleCloseAfterMessages(event: CloseEvent): Promise<void> {
        await pendingMessageDispatch.catch(reportMessageError);
        emitReadyState();
        handlers.onClose?.(event, socket);
        cleanup();
    }

    function reportMessageError(error: unknown): void {
        try {
            handlers.onMessageError?.(error, socket);
        } catch {
            // Keep the transport queue draining even if a consumer error handler fails.
        }
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
