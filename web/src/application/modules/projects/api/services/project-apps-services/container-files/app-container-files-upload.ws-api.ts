import { BaseWebSocketApi, type WebSocketSubscription, toWebSocketError } from "@infrastructure/websocket";
import { Err, type Result } from "oxide.ts";

import { session } from "@infrastructure/api";

import { CancelException } from "@infrastructure/exceptions/cancel";
import { NetworkException } from "@infrastructure/exceptions/network";

import type {
    AppContainerFiles_UploadOne_Req,
    AppContainerFiles_UploadOne_Res,
} from "./app-container-files.api.contracts";
import type { AppContainerFilesApiValidator } from "./app-container-files.api.validator";

// A piece of the file a message carries; the server takes up to 4 MiB.
const PIECE_SIZE = 256 * 1024;
// How much may wait in the socket's buffer before the next piece is read: a
// browser's websocket buffers all it is given, the whole file if let.
const BUFFERED_MAX = 4 * PIECE_SIZE;
const BUFFER_POLL_MS = 50;

const END_MESSAGE = JSON.stringify({ type: "end" });

function wait(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Uploads a file into an app's container over a websocket: the file in pieces,
 * then an end; the server answers done or the error. No proxy's timeout on a
 * request's body cuts it, as one that takes minutes would be.
 */
export class AppContainerFilesUploadWsApi extends BaseWebSocketApi {
    constructor(private readonly validator: AppContainerFilesApiValidator) {
        super();
    }

    uploadOne(
        req: AppContainerFiles_UploadOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppContainerFiles_UploadOne_Res, Error>> {
        const accessToken = session.getToken();

        if (!accessToken) {
            return Promise.resolve(Err(new Error("Access token not found.")));
        }

        const { projectID, env, appID, nodeId, containerId, path, file, extract, compressionFormat, overwrite } =
            req.data;
        const { onProgress } = req.data;

        return new Promise(resolve => {
            let settled = false;
            let subscription: WebSocketSubscription | undefined;

            const settle = (result: Result<AppContainerFiles_UploadOne_Res, Error>) => {
                if (settled) {
                    return;
                }

                settled = true;
                resolve(result);
                subscription?.close();
            };

            const isSending = (socket: WebSocket) => !settled && socket.readyState === WebSocket.OPEN;

            // What has left the browser: given to the socket, and no longer in its buffer.
            const report = (socket: WebSocket, given: number) => {
                onProgress?.(Math.max(0, Math.min(file.size, given - socket.bufferedAmount)), file.size);
            };

            const send = async (socket: WebSocket) => {
                let given = 0;

                try {
                    while (given < file.size) {
                        while (isSending(socket) && socket.bufferedAmount > BUFFERED_MAX) {
                            report(socket, given);
                            await wait(BUFFER_POLL_MS);
                        }

                        const piece = await file.slice(given, given + PIECE_SIZE).arrayBuffer();

                        // The server may have answered meanwhile - an error, as it refused.
                        if (!isSending(socket)) {
                            return;
                        }

                        socket.send(piece);
                        given += piece.byteLength;
                        report(socket, given);
                    }

                    if (!isSending(socket)) {
                        return;
                    }

                    socket.send(END_MESSAGE);
                    given += END_MESSAGE.length;

                    while (isSending(socket) && socket.bufferedAmount > 0) {
                        report(socket, given);
                        await wait(BUFFER_POLL_MS);
                    }

                    report(socket, given);
                } catch (error) {
                    // The file could not be read: changed, or gone, since it was picked.
                    settle(Err(error instanceof Error ? error : new Error("Failed to read the file.")));
                }
            };

            try {
                const url = this.client.buildUrl(
                    `projects/${encodeURIComponent(projectID)}/${encodeURIComponent(env)}/apps/${encodeURIComponent(appID)}/container/file-upload/stream`,
                    {
                        nodeId,
                        containerId,
                        path,
                        extract,
                        compressionFormat,
                        overwrite,
                        fileName: file.name,
                        fileSize: file.size,
                    },
                );

                subscription = this.client.connect(
                    url,
                    {
                        onOpen: (_event, socket) => {
                            void send(socket);
                        },
                        onMessage: message => {
                            settle(this.validator.uploadAnswer(message));
                        },
                        onClose: () => {
                            // A refusal before the stream opened - the session, a permission -
                            // comes as a close the browser does not explain.
                            settle(
                                Err(
                                    signal?.aborted
                                        ? new CancelException("The upload was cancelled.")
                                        : new NetworkException("The connection closed before the server answered."),
                                ),
                            );
                        },
                    },
                    {
                        signal,
                        closeOnError: true,
                        protocols: ["access_token", accessToken],
                    },
                );
            } catch (error) {
                settle(Err(toWebSocketError(error, "Failed to upload the file.")));
            }
        });
    }
}
