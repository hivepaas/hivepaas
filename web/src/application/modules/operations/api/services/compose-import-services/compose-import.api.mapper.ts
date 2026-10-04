import type { ComposeFileInput, ComposeImportBody } from "~/operations/domain";

import type { ComposeImportWireBody } from "./compose-import.api.contracts";

/** Base64 of bytes, which is how JSON carries a Go []byte. */
function toBase64(bytes: Uint8Array): string {
    const chunkSize = 0x8000;
    let binary = "";
    for (let index = 0; index < bytes.length; index += chunkSize) {
        binary += String.fromCharCode(...bytes.subarray(index, index + chunkSize));
    }

    return window.btoa(binary);
}

async function encodeFile(file: ComposeFileInput): Promise<string> {
    if (typeof file === "string") {
        return toBase64(new TextEncoder().encode(file));
    }

    return toBase64(new Uint8Array(await file.arrayBuffer()));
}

class BodyMapper {
    async toApi(
        body: ComposeImportBody,
        extra: Pick<ComposeImportWireBody, "planHash" | "acceptIssues"> = {},
    ): Promise<ComposeImportWireBody> {
        const files: Record<string, string> = {};
        for (const [path, file] of Object.entries(body.files)) {
            files[path] = await encodeFile(file);
        }

        return {
            compose: body.compose,
            dotEnv: body.dotEnv,
            files,
            variables: body.variables,
            project: body.project,
            profiles: body.profiles,
            services: body.services,
            volume: body.volumeId ? { id: body.volumeId } : undefined,
            selection: body.selection,
            deploy: body.deploy,
            ...extra,
        };
    }
}

export class ComposeImportApiMapper {
    readonly body = new BodyMapper();
}
