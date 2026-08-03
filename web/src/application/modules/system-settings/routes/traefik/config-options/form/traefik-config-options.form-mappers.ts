import type { TraefikConfigOptions, TraefikConfigOptions_UpdateOne_Payload } from "~/system-settings/api/services";

import type { TraefikConfigOptionsFormInput } from "../schemas";

export function mapTraefikConfigOptionsToFormInput(configOptions: TraefikConfigOptions): TraefikConfigOptionsFormInput {
    const { startupCommand } = configOptions;

    return {
        startupCommand: {
            logLevel: startupCommand.logLevel || "",
            accessLog: startupCommand.accessLog,
            http3: startupCommand.http3,
            fastProxy: startupCommand.fastProxy,
            argsText: startupCommand.args.join("\n"),
        },
    };
}

export function mapFormOutputToUpdatePayload(values: {
    startupCommand: {
        logLevel: string;
        accessLog: boolean;
        http3: boolean;
        fastProxy: boolean;
        argsText: string;
    };
}): TraefikConfigOptions_UpdateOne_Payload {
    const args = values.startupCommand.argsText
        .split("\n")
        .map(line => line.trim())
        .filter(Boolean);

    return {
        startupCommand: {
            logLevel: values.startupCommand.logLevel,
            accessLog: values.startupCommand.accessLog,
            http3: values.startupCommand.http3,
            fastProxy: values.startupCommand.fastProxy,
            args,
        },
    };
}
