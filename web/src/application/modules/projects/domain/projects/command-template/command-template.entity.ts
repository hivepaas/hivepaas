import type { EAppScheduledJobArgSeparator } from "~/projects/module-shared/enums";

import type { ECommandTemplateKind, ESettingStatus } from "@application/shared/enums";

export interface ProjectCommandTemplateEnvVar {
    key: string;
    value: string;
    isLiteral: boolean;
}

export interface ProjectCommandTemplateArg {
    use: boolean;
    name: string;
    value: string;
}

export interface ProjectCommandTemplateArgGroup {
    enabled: boolean;
    exportEnv: string;
    separator: EAppScheduledJobArgSeparator;
    args: ProjectCommandTemplateArg[];
}

export interface ProjectCommandTemplateConsoleSize {
    width: number;
    height: number;
}

export const PROJECT_COMMAND_TEMPLATE_DEFAULT_CONSOLE_SIZE = {
    width: 120,
    height: 40,
} as const satisfies ProjectCommandTemplateConsoleSize;

export interface ProjectCommandTemplate {
    id: string;
    type: "command-template";
    name: string;
    kind: ECommandTemplateKind;
    status: ESettingStatus;
    inherited: boolean;
    availableInProjects: boolean;
    default: boolean;
    updateVer: number;
    createdAt: Date;
    updatedAt: Date | null;
    expireAt: Date | null;
    command: string;
    script: string;
    workingDir: string;
    envVars: ProjectCommandTemplateEnvVar[];
    argGroups: ProjectCommandTemplateArgGroup[];
    consoleSize: ProjectCommandTemplateConsoleSize;
    tty: boolean;
    size: number;
}
