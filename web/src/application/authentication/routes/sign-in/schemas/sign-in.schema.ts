import z from "zod";

export const SignInSchema = z.object({
  email: z.string().trim().min(1, "Email Address is required").email(),
  password: z.string().trim().min(1, "Password is required"),
  rememberMe: z.boolean(),
});

export type SignInSchemaInput = z.input<typeof SignInSchema>;
export type SignInSchemaOutput = z.output<typeof SignInSchema>;