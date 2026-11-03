import type { SettingUsage } from "@application/shared/utils";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

/**
 * Addressed by the path of the delete that was refused rather than by a setting
 * id, because settings are reached through a route per type - the usages of one
 * hang off the same path its delete was sent to. That is also why this takes a
 * URL: the caller is a shared dialog reacting to a refusal, and the only thing it
 * knows about the setting is where the request went.
 */
export type SettingUsages_FindMany_Req = ApiRequestBase<{ deleteUrl: string }>;
export type SettingUsages_FindMany_Res = ApiResponseBase<SettingUsage[]>;
