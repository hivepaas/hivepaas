export interface ProfileApiKey {
    id: string;
    name: string;
    key: string;
    accessAction: {
        read: boolean;
        write: boolean;
        delete: boolean;
    };
    expireAt: Date;
    status: string;
}
