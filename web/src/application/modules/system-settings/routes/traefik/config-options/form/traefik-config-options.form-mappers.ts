import type { TraefikConfigOptions, TraefikConfigOptions_UpdateOne_Payload } from "~/system-settings/api/services";
import { CONFIRM_WINDOW } from "~/system-settings/module-shared/utils";

import type { TraefikConfigOptionsFormInput } from "../schemas";

export function mapTraefikConfigOptionsToFormInput(configOptions: TraefikConfigOptions): TraefikConfigOptionsFormInput {
    const { startupCommand } = configOptions;

    return {
        startupCommand: {
            logLevel: startupCommand.logLevel || "",
            accessLog: startupCommand.accessLog,
            http3: startupCommand.http3,
            fastProxy: startupCommand.fastProxy,
            openPortsText: (startupCommand.openPorts ?? []).join(", "),
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
        // Asking for the same window as the other trials. Traefik takes longer to
        // come back than a routing change does, and the server's floor accounts
        // for that, but the operator should not be given a shorter budget here
        // than anywhere else.
        confirmWindow: CONFIRM_WINDOW,
    };
}
