import type { DeviceInfo } from "@infrastructure/device";
import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import {
    type AuthApiValidator,
    type Auth_ForgotPassword_Req,
    type Auth_ForgotPassword_Res,
    type Auth_GetLoginOptions_Res,
    type Auth_ResetPassword_Req,
    type Auth_ResetPassword_Res,
    type Auth_Send2FAToken_Req,
    type Auth_Send2FAToken_Res,
    type Auth_SignIn2FA_Req,
    type Auth_SignIn2FA_Res,
    type Auth_SignIn_Req,
    type Auth_SignIn_Res,
    type Auth_SignUp_Req,
    type Auth_SignUp_Res,
    type Auth_ValidateInviteToken_Req,
    type Auth_ValidateInviteToken_Res,
    type Auth_ValidateResetToken_Req,
    type Auth_ValidateResetToken_Res,
} from "@application/authentication/api/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class AuthApi extends BaseApi {
    public constructor(
        private readonly validator: AuthApiValidator,
        private readonly device: DeviceInfo,
    ) {
        super();
    }

    /**
     * Sign up
     */
    async signUp(request: Auth_SignUp_Req): Promise<Result<Auth_SignUp_Res, Error>> {
        // const json = this.mapper.signUp.toEmailApi(request.data);
        const { inviteToken, data } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/users/signup-complete", {
                    inviteToken,
                    password: data.password,
                    fullName: data.fullName,
                    photo: data.photo ?? {
                        fileName: "",
                        dataBase64: "",
                    },
                    mfaTotpSecret: data.mfaTotpSecret == "" ? undefined : data.mfaTotpSecret,
                    passcode: data.passcode == "" ? undefined : data.passcode,
                }),
            ).pipe(
                map(this.validator.signUp),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Sign in with email and password
     */
    async signIn(request: Auth_SignIn_Req): Promise<Result<Auth_SignIn_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.post(
                    "/auth/login-with-password",
                    {
                        username: request.data.email,
                        password: request.data.password,
                        trustedDeviceId: request.data.isTrustDevice ? await this.device.getFingerprint() : undefined,
                    },
                    // {
                    //     withCredentials: true,
                    // },
                ),
            ).pipe(
                map(this.validator.signIn),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Send 2-factor authentication token
     */
    async send2FAToken(request: Auth_Send2FAToken_Req): Promise<Result<Auth_Send2FAToken_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.post("/auth/send-mfa-passcode", {
                    mfaToken: request.data.mfaToken,
                }),
            ).pipe(
                map(this.validator.send2FAToken),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Sign in with 2-factor authentication
     */
    async signIn2FA(request: Auth_SignIn2FA_Req): Promise<Result<Auth_SignIn2FA_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.post(
                    "/auth/login-with-passcode",
                    {
                        passcode: request.data.code,
                        mfaToken: request.data.mfaToken,
                    },
                    // {
                    //     withCredentials: true,
                    // },
                ),
            ).pipe(
                map(this.validator.signIn2FA),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Forgot password
     */
    async forgotPassword(request: Auth_ForgotPassword_Req): Promise<Result<Auth_ForgotPassword_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.post("/users/password-forgot", {
                    email: request.data.email,
                }),
            ).pipe(
                map(this.validator.forgotPassword),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Validate reset token
     */
    async validateResetToken(
        request: Auth_ValidateResetToken_Req,
    ): Promise<Result<Auth_ValidateResetToken_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.post("/users/password-reset-token-validate", {
                    email: request.data.email,
                    token: request.data.token,
                }),
            ).pipe(
                map(this.validator.validateResetToken),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Reset password
     */
    async resetPassword(request: Auth_ResetPassword_Req): Promise<Result<Auth_ResetPassword_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.post("/users/password-reset", {
                    email: request.data.email,
                    token: request.data.token,
                    newPassword: request.data.newPassword,
                }),
            ).pipe(
                map(this.validator.resetPassword),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Get login options
     */
    async getLoginOptions(signal: AbortSignal): Promise<Result<Auth_GetLoginOptions_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/auth/login-options", { signal })).pipe(
                map(this.validator.getLoginOptions),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Validate invite token
     */
    async validateInviteToken(
        request: Auth_ValidateInviteToken_Req,
    ): Promise<Result<Auth_ValidateInviteToken_Res, Error>> {
        const { data } = request;

        return lastValueFrom(
            from(
                this.client.v1.post("/users/signup-begin", {
                    inviteToken: data.inviteToken,
                }),
            ).pipe(
                map(this.validator.validateInviteToken),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
