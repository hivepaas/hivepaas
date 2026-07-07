import {
    PROJECT_COMMAND_TEMPLATE_DEFAULT_CONSOLE_SIZE,
    type ProjectCommandTemplate,
    type ProjectCommandTemplateArgGroup,
    type ProjectCommandTemplateEnvVar,
} from "~/projects/domain";

import {
    PROJECT_COMMAND_TEMPLATE_COMMAND_MODE,
    type ProjectCommandTemplateFormInput,
    type ProjectCommandTemplateFormOutput,
} from "../schemas";

export function createEmptyProjectCommandTemplateFormDefaults(): ProjectCommandTemplateFormInput {
    return {
        name: "",
        kind: "",
        commandMode: PROJECT_COMMAND_TEMPLATE_COMMAND_MODE.Command,
        command: "",
        script: "",
        workingDir: "",
        tty: false,
        consoleSize: { ...PROJECT_COMMAND_TEMPLATE_DEFAULT_CONSOLE_SIZE },
        envVars: [],
        argGroups: [],
        default: false,
    };
}

export function mapProjectCommandTemplateToFormInput(
    commandTemplate: ProjectCommandTemplate,
): ProjectCommandTemplateFormInput {
    const { script } = commandTemplate;

    return {
        name: commandTemplate.name,
        kind: commandTemplate.kind,
        commandMode: script.trim()
            ? PROJECT_COMMAND_TEMPLATE_COMMAND_MODE.Script
            : PROJECT_COMMAND_TEMPLATE_COMMAND_MODE.Command,
        command: commandTemplate.command,
        script,
        workingDir: commandTemplate.workingDir,
        tty: commandTemplate.tty,
        consoleSize: { ...commandTemplate.consoleSize },
        envVars: commandTemplate.envVars,
        argGroups: commandTemplate.argGroups,
        default: commandTemplate.default,
    };
}

export function normalizeProjectCommandTemplateEnvVars(
    envVars: ProjectCommandTemplateFormOutput["envVars"],
): ProjectCommandTemplateEnvVar[] {
    return envVars.reduce<ProjectCommandTemplateEnvVar[]>((acc, item) => {
        const key = item.key.trim();

        if (!key) {
            return acc;
        }

        return [
            ...acc,
            {
                key,
                value: item.value,
                isLiteral: item.isLiteral ?? false,
            },
        ];
    }, []);
}

export function normalizeProjectCommandTemplateArgGroups(
    argGroups: ProjectCommandTemplateFormOutput["argGroups"],
): ProjectCommandTemplateArgGroup[] {
    return argGroups.map(group => ({
        enabled: group.enabled,
        exportEnv: group.exportEnv.trim(),
        separator: group.separator,
        args: group.args.map(arg => ({
            use: arg.use,
            name: arg.name.trim(),
            value: arg.value,
        })),
    }));
}
