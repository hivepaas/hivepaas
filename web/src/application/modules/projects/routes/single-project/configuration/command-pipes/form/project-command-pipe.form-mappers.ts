import type { ProjectCommandPipe } from "~/projects/domain";

import type { ProjectCommandPipeFormInput } from "../schemas";

export function createEmptyProjectCommandPipeFormDefaults(): ProjectCommandPipeFormInput {
    return {
        name: "",
        sourceCommandId: "",
        targetCommandId: "",
        inheritable: true,
        default: false,
    };
}

export function mapProjectCommandPipeToFormInput(commandPipe: ProjectCommandPipe): ProjectCommandPipeFormInput {
    return {
        name: commandPipe.name,
        sourceCommandId: commandPipe.sourceCommand?.id ?? "",
        targetCommandId: commandPipe.targetCommand?.id ?? "",
        inheritable: commandPipe.inheritable,
        default: commandPipe.default,
    };
}
