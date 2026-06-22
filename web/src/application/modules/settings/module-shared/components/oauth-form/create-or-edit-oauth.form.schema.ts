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
    default: z.boolean(),
});

export type CreateOrEditOAuthFormInput = z.input<typeof CreateOrEditOAuthFormSchema>;
export type CreateOrEditOAuthFormOutput = z.output<typeof CreateOrEditOAuthFormSchema>;
