import { z } from "zod";
import { ERegistryAuthKind } from "~/settings/domain";

/**
 * An Amazon ECR registry's addresses, as the API reads its region from them: the
 * API checks the address, this only shows the region.
 */
const ECR_ADDRESS_REGEXES = [
    /^([0-9]{12})\.dkr\.ecr(?:-fips)?\.([a-z]{2}(?:-[a-z]+)+-[0-9])\.amazonaws\.com(?:\.cn)?$/,
    /^([0-9]{12})\.dkr-ecr\.([a-z]{2}(?:-[a-z]+)+-[0-9])\.on\.aws$/,
];

/** The region of an Amazon ECR registry's address, or "" for any other address. */
export function ecrRegionOf(address: string): string {
    const value = address.trim().toLowerCase();
    for (const regex of ECR_ADDRESS_REGEXES) {
        const region = regex.exec(value)?.[2];
        if (region) {
            return region;
        }
    }
    return "";
}

export const CreateOrEditRegistryAuthFormSchema = z
    .object({
        name: z.string().trim().min(1, "Name is required"),
        kind: z.enum([ERegistryAuthKind.Basic, ERegistryAuthKind.AwsEcr]),
        address: z.string().trim().min(1, "Server address is required"),
        username: z.string().trim(),
        // 8 KB, the API's: a Google Artifact Registry JSON key in base64 is some 3 KB.
        password: z.string().max(8 * 1024, "At most 8 KB"),
        ecrKeyAuth: z.object({ id: z.string(), name: z.string() }).nullable(),
        ecrRoleArn: z.string().trim(),
        readonly: z.boolean(),
        inheritable: z.boolean(),
        default: z.boolean(),
    })
    .superRefine((values, ctx) => {
        if (values.kind === ERegistryAuthKind.AwsEcr) {
            if (!values.ecrKeyAuth?.id) {
                ctx.addIssue({ code: "custom", path: ["ecrKeyAuth"], message: "Key auth is required" });
            }
            return;
        }

        if (!values.username) {
            ctx.addIssue({ code: "custom", path: ["username"], message: "Username is required" });
        }
        if (!values.password) {
            ctx.addIssue({ code: "custom", path: ["password"], message: "Password is required" });
        }
    });

export type CreateOrEditRegistryAuthFormInput = z.input<typeof CreateOrEditRegistryAuthFormSchema>;
export type CreateOrEditRegistryAuthFormOutput = z.output<typeof CreateOrEditRegistryAuthFormSchema>;
