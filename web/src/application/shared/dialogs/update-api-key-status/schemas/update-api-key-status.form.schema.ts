import { z } from "zod";

import { EProfileApiKeyStatus } from "@application/shared/enums";

export const UpdateApiKeyStatusFormSchema = z.object({
    status: z.nativeEnum(EProfileApiKeyStatus),
    expireAt: z
        .date()
        .refine(
            date => {
                const now = new Date();
                now.setSeconds(0, 0);
                return date > now;
            },
            { message: "Expiration date must be in the future" },
        )
        .optional(),
});

export type UpdateApiKeyStatusFormInput = z.input<typeof UpdateApiKeyStatusFormSchema>;
export type UpdateApiKeyStatusFormOutput = z.output<typeof UpdateApiKeyStatusFormSchema>;
