import type { SettingsPendingChange } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export interface TraefikConfigOptionsStartupCommand {
    logLevel: string;
    accessLog: boolean;
    http3: boolean;
    fastProxy: boolean;
    openPorts?: string[];
    args: string[];
}

export interface TraefikConfigOptions {
    startupCommand: TraefikConfigOptionsStartupCommand;

    /**
     * The trial in progress, if there is one.
     *
     * This is the only reliable way to find out about it. Applying these options
     * replaces Traefik's task, so the response to the update that started the
     * trial travels down a connection that is being cut and often never arrives.
     */
    pendingChange: SettingsPendingChange | null;
}

export type TraefikConfigOptions_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type TraefikConfigOptions_FindOne_Res = ApiResponseBase<TraefikConfigOptions>;

export type TraefikConfigOptions_UpdateOne_Payload = {
    startupCommand: TraefikConfigOptionsStartupCommand;

    /**
     * How long the change may stay unconfirmed before it is undone, as a Go
     * duration string ("5m"). Clamped server-side; there is no way to ask for no
     * trial at all.
     */
    confirmWindow?: string;
};

export type TraefikConfigOptions_UpdateOne_Req = ApiRequestBase<{
    payload: TraefikConfigOptions_UpdateOne_Payload;
}>;
// Null when the request asked for nothing Traefik is not already running: there
// is no restart to survive, so there is nothing to confirm.
export type TraefikConfigOptions_UpdateOne_Res = ApiResponseBase<{
    pendingChange: SettingsPendingChange | null;
}>;

export type TraefikConfigOptions_ConfirmChange_Req = ApiRequestBase<{ changeId: string }>;
export type TraefikConfigOptions_ConfirmChange_Res = ApiResponseBase<{ type: "success" }>;

export type TraefikConfigOptions_RevertChange_Req = ApiRequestBase<{ changeId: string }>;
export type TraefikConfigOptions_RevertChange_Res = ApiResponseBase<{
    reverted: boolean;
    reason: string | null;
}>;
