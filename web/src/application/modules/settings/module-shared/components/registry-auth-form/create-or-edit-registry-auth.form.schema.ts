import { z } from "zod";
import { ERegistryAuthKind } from "~/settings/domain";

/** An Amazon ECR registry's address, as the API reads its account and region from it. */
export const ECR_ADDRESS_REGEX =
    /^([0-9]{12})\.dkr\.ecr(?:-fips)?\.([a-z]{2}(?:-[a-z]+)+-[0-9])\.amazonaws\.com(?:\.cn)?$/;
const AWS_ROLE_ARN_REGEX = /^arn:aws(-cn|-us-gov)?:iam::[0-9]{12}:role\/[\w+=,.@/-]+$/;

/** The region of an Amazon ECR registry's address, or "" for any other address. */
export function ecrRegionOf(address: string): string {
    return ECR_ADDRESS_REGEX.exec(address.trim().toLowerCase())?.[2] ?? "";
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
            if (!ECR_ADDRESS_REGEX.test(values.address.toLowerCase())) {
                ctx.addIssue({
                    code: "custom",
                    path: ["address"],
                    message: "An Amazon ECR registry: <account>.dkr.ecr.<region>.amazonaws.com",
                });
            }
            if (!values.ecrKeyAuth?.id) {
                ctx.addIssue({ code: "custom", path: ["ecrKeyAuth"], message: "Key auth is required" });
            }
            if (values.ecrRoleArn && !AWS_ROLE_ARN_REGEX.test(values.ecrRoleArn)) {
                ctx.addIssue({
                    code: "custom",
                    path: ["ecrRoleArn"],
                    message: "An IAM role's ARN: arn:aws:iam::<account>:role/<name>",
                });
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
