interface Photo {
    fileName: string;
    dataBase64: string;
}

interface Data {
    email: string;
    username: string;
    fullName: string;
    position: string;
    password: string;
    photo: Photo | null;
    mfaTotpSecret?: string;
    passcode?: string;
}

export type SignUp = {
    inviteToken: string;
    data: Omit<Data, "email">;
};
