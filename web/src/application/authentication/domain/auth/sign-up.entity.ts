interface Photo {
    fileName: string;
    dataBase64: string;
}

interface Data {
    email: string;
    fullName: string;
    password: string;
    photo: Photo | null;
    mfaTotpSecret?: string;
    passcode?: string;
}

export type SignUp = {
    inviteToken: string;
    data: Omit<Data, "email">;
};
