import { type SignUp } from "@application/authentication/domain";

interface RequestData {
    email?: string | null;
    fullName?: string;
    password: string;
    photo: {
        fileName: string;
        dataBase64: string;
    } | null;
}

class SignUpMapper {
    private map = (data: SignUp["data"]): Omit<RequestData, "email"> => {
        return {
            password: data.password,
            photo: data.photo,
        };
    };

    // toEmailApi = (domain: SignUp): Omit<RequestData, "email"> => {
    //     if (domain.type !== ESignUpType.Email) {
    //         throw new Error("Invalid sign up type");
    //     }

    //     return this.map(domain.data);
    // };

    // toCodeApi = (domain: SignUp): RequestData => {
    //     if (domain.type !== ESignUpType.Code) {
    //         throw new Error("Invalid sign up type");
    //     }

    //     return {
    //         ...this.map(domain.data),
    //         email: domain.data.email,
    //     };
    // };
}

export class AuthApiMapper {
    signUp = new SignUpMapper();
}
