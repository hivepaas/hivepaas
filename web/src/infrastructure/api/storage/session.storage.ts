const KEY = "token";

class SessionStorage {
    hasToken(): boolean {
        const token = this.getToken();

        return token !== null;
    }

    getToken(): string | null {
        const token = (localStorage.getItem(KEY) ?? "").trim();

        if (token === "") {
            return null;
        }

        return token;
    }

    setToken(token: string): void {
        const t = token.trim();

        if (t === "") {
            throw new Error("Token should not be empty string");
        }

        localStorage.setItem(KEY, t);
    }

    removeToken(): void {
        localStorage.removeItem(KEY);
    }
}

export const session = new SessionStorage();
