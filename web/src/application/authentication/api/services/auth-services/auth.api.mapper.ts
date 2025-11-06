import { type SignUp } from "@application/authentication/domain";
import { ESignUpType } from "@application/authentication/enums";

interface RequestData {
    email: string;
    firstName: string;
    lastName: string;
    position: string;
    entity: { id: string } | null;
    timezone: string;
    mobilePhone: string;
    officePhone: string;
    password: string;
    photo: {
        fileName: string;
        dataBase64: string;
    } | null;
}

class SignUpMapper {
    private map = (data: SignUp["data"]): Omit<RequestData, "email"> => {
        return {
            firstName: data.firstName,
            lastName: data.lastName,
            position: data.position,
            entity: data.entity ? { id: data.entity.id } : null,
            timezone: data.timezone,
            mobilePhone: data.mobilePhone,
            officePhone: data.officePhone,
            password: data.password,
            photo: data.photo,
        };
    };

    toEmailApi = (domain: SignUp): Omit<RequestData, "email"> => {
        if (domain.type !== ESignUpType.Email) {
            throw new Error("Invalid sign up type");
        }

        return this.map(domain.data);
    };

    toCodeApi = (domain: SignUp): RequestData => {
        if (domain.type !== ESignUpType.Code) {
            throw new Error("Invalid sign up type");
        }

        return {
            ...this.map(domain.data),
            email: domain.data.email,
        };
    };
}

export class AuthApiMapper {
    signUp = new SignUpMapper();
}
