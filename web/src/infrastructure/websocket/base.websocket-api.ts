import { type WebSocketClient, createWebSocketClient } from "./client";

const wc = createWebSocketClient();

export abstract class BaseWebSocketApi {
    constructor(client: WebSocketClient = wc) {
        this.client = client;
    }

    protected readonly client: WebSocketClient;
}
