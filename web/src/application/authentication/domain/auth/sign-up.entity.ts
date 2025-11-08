interface Entity {
    id: string;
    name: string;
}

interface Photo {
    fileName: string;
    dataBase64: string;
}

interface Data {
    email: string;
    firstName: string;
    lastName: string;
    position: string;
    entity?: Entity | null;
    timezone: string;
    mobilePhone: string;
    officePhone: string;
    password: string;
    photo: Photo | null;
}

export type SignUp = {
    inviteToken: string;
    data: Omit<Data, "email">;
};
