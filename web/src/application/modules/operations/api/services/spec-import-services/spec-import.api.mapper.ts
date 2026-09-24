import type { SpecImportScope } from "~/operations/domain";

import type { SpecImportBody, SpecImportWireBody } from "./spec-import.api.contracts";

/**
 * The bundle is read once per file, however often the plan is asked for again
 * while the operator changes the selection.
 */
const encoded = new WeakMap<File, Promise<string>>();

/** Base64 of a file's bytes, which is how JSON carries a Go []byte. */
function encodeFile(file: File): Promise<string> {
    const cached = encoded.get(file);
    if (cached) {
        return cached;
    }
    const promise = file.arrayBuffer().then(buffer => {
        const bytes = new Uint8Array(buffer);
        const chunkSize = 0x8000;
        let binary = "";

        for (let index = 0; index < bytes.length; index += chunkSize) {
            binary += String.fromCharCode(...bytes.subarray(index, index + chunkSize));
        }

        return window.btoa(binary);
    });
    encoded.set(file, promise);

    return promise;
}

export function resolveSpecImportEndpoint(scope: SpecImportScope, step: "validate" | "apply"): string {
    switch (scope.type) {
        case "project":
            return `/projects/${scope.projectID}/spec/import/${step}`;
        case "project-env":
            return `/projects/${scope.projectID}/${scope.projectEnvID}/spec/import/${step}`;
        case "global":
        default:
            return `/spec/import/${step}`;
    }
}

class BodyMapper {
    async toApi(
        body: SpecImportBody,
        extra: Pick<SpecImportWireBody, "planHash" | "acceptIssues"> = {},
    ): Promise<SpecImportWireBody> {
        return {
            bundle: await encodeFile(body.bundle),
            // An empty passphrase is no passphrase: the server asks for one when it needs it.
            passphrase: body.passphrase === "" ? undefined : body.passphrase,
            selection: body.selection,
            options: body.options,
            ...extra,
        };
    }
}

export class SpecImportApiMapper {
    readonly body = new BodyMapper();
}
