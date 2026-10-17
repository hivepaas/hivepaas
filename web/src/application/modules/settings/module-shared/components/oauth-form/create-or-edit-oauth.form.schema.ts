import { z } from "zod";

import { EOAuthKind } from "@application/shared/enums";

export const CreateOrEditOAuthFormSchema = z.object({
    name: z.string(),
    kind: z.nativeEnum(EOAuthKind),
    organization: z.string().min(1, "Organization is required"),
    clientId: z.string().min(1, "Client ID is required"),
    clientSecret: z.string().min(1, "Client Secret is required"),
    authURL: z.string(),
    tokenURL: z.string(),
    profileURL: z.string(),
    autoDiscoveryURL: z.string(),
    scopes: z.string(),
    trustEmail: z.boolean(),
    default: z.boolean(),
});

/** A GitLab or a Gitea of one's own is signed in with by its URLs; every other kind has its provider's own. */
export function hasCustomURLs(kind: EOAuthKind): boolean {
    return kind === EOAuthKind.Gitlab || kind === EOAuthKind.Gitea;
}

export type CreateOrEditOAuthFormInput = z.input<typeof CreateOrEditOAuthFormSchema>;
export type CreateOrEditOAuthFormOutput = z.output<typeof CreateOrEditOAuthFormSchema>;
