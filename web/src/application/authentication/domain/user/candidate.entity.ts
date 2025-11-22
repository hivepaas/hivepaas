import { type ESecuritySettings, type EUserRole } from "@application/shared/enums";

export type Candidate =
    | {
          role: EUserRole;
          email: string;
          securityOption: typeof ESecuritySettings.PasswordOnly;
          accessExpiration: Date | null;
      }
    | {
          role: EUserRole;
          email: string;
          securityOption: typeof ESecuritySettings.EnforceSSO;
          accessExpiration: Date | null;
      }
    | {
          role: EUserRole;
          email: string;
          securityOption: typeof ESecuritySettings.Password2FA;
          accessExpiration: Date | null;
          mfaTotpSecret: string;
          qrCode: string;
          secretKey: string;
      };
