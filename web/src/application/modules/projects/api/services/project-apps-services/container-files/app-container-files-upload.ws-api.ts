import { BaseWebSocketApi, type WebSocketSubscription, toWebSocketError } from "@infrastructure/websocket";
import { Err, type Result } from "oxide.ts";

import { session } from "@infrastructure/api";

import { CancelException } from "@infrastructure/exceptions/cancel";
import { NetworkException } from "@infrastructure/exceptions/network";

import type { AppContainerFilesApi } from "./app-container-files.api";
import {
    type AppContainerFiles_UploadOne_Req,
    type AppContainerFiles_UploadOne_Res,
    appContainerFilesUploadQuery,
} from "./app-container-files.api.contracts";
import type { AppContainerFilesApiValidator } from "./app-container-files.api.validator";

// A piece of the file a message carries; the server takes up to 4 MiB.
const PIECE_SIZE = 256 * 1024;
// How much may be on its way - sent, and not yet taken by the copy into the
// container. The server says what it has taken after each piece: the sender
// waits on that, not on a timer, which a tab in the background runs once a
// second, or once a minute.
const IN_FLIGHT_MAX = 8 * 1024 * 1024;
// How often the progress is told at most, for a dialog to draw.
const PROGRESS_EVERY_MS = 100;

const END_MESSAGE = JSON.stringify({ type: "end" });

/**
 * Uploads a file into an app's container over a websocket: the file in pieces,
 * then an end; the server answers done or the error. No proxy's timeout on a
 * request's body cuts it, as one that takes minutes would be.
 */
export class AppContainerFilesUploadWsApi extends BaseWebSocketApi {
    constructor(
        private readonly validator: AppContainerFilesApiValidator,
        private readonly filesApi: AppContainerFilesApi,
    ) {
        super();
    }

    async uploadOne(
        req: AppContainerFiles_UploadOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppContainerFiles_UploadOne_Res, Error>> {
        const checked = await this.filesApi.checkUpload(req, signal);

        if (checked.isErr()) {
            return Err(checked.unwrapErr());
        }

        return this.stream(req, signal);
    }

    private stream(
        req: AppContainerFiles_UploadOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppContainerFiles_UploadOne_Res, Error>> {
        // Read after the check, which refreshed it if it had expired.
        const accessToken = session.getToken();

        if (!accessToken) {
            return Promise.resolve(Err(new Error("Access token not found.")));
        }

        const { projectID, env, appID, file, onProgress } = req.data;

        return new Promise(resolve => {
            let settled = false;
            let subscription: WebSocketSubscription | undefined;
            // What the copy has taken, as the server last said.
            let received = 0;
            let toldAt = 0;
            let wakeSender: (() => void) | undefined;

            const wake = () => {
                const resume = wakeSender;
                wakeSender = undefined;
                resume?.();
            };

            const settle = (result: Result<AppContainerFiles_UploadOne_Res, Error>) => {
                if (settled) {
                    return;
                }

                settled = true;
                signal?.removeEventListener("abort", cancelled);
                resolve(result);
                wake();
                subscription?.close();
            };

            // Cancelled is done with at once: the socket closes after what it
            // buffered is sent, which on a slow link takes seconds.
            const cancelled = () => {
                settle(Err(new CancelException("The upload was cancelled.")));
            };
            signal?.addEventListener("abort", cancelled, { once: true });

            const tell = (force = false) => {
                const now = performance.now();

                if (settled || (!force && now - toldAt < PROGRESS_EVERY_MS)) {
                    return;
                }

                toldAt = now;
                onProgress?.(Math.min(received, file.size), file.size);
            };

            const isSending = (socket: WebSocket) => !settled && socket.readyState === WebSocket.OPEN;

            const send = async (socket: WebSocket) => {
                let given = 0;

                try {
                    while (given < file.size) {
                        while (isSending(socket) && given - received >= IN_FLIGHT_MAX) {
                            await new Promise<void>(resume => {
                                wakeSender = resume;
                            });
                        }

                        const piece = await file.slice(given, given + PIECE_SIZE).arrayBuffer();

                        // The server may have answered meanwhile - an error, as it refused.
                        if (!isSending(socket)) {
                            return;
                        }

                        socket.send(piece);
                        given += piece.byteLength;
                    }

                    if (isSending(socket)) {
                        socket.send(END_MESSAGE);
                    }
                } catch (error) {
                    // The file could not be read: changed, or gone, since it was picked.
                    settle(Err(error instanceof Error ? error : new Error("Failed to read the file.")));
                }
            };

            try {
                const url = this.client.buildUrl(
                    `projects/${encodeURIComponent(projectID)}/${encodeURIComponent(env)}/apps/${encodeURIComponent(appID)}/container/file-upload/stream`,
                    appContainerFilesUploadQuery(req.data),
                );

                subscription = this.client.connect(
                    url,
                    {
                        onOpen: (_event, socket) => {
                            tell(true);
                            void send(socket);
                        },
                        onMessage: message => {
                            const parsed = this.validator.uploadMessage(message);

                            if (parsed.type === "progress") {
                                ({ received } = parsed);
                                tell(received >= file.size);
                                wake();
                                return;
                            }

                            settle(parsed.result);
                        },
                        onClose: () => {
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
