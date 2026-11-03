import FingerprintJS, { type Agent, type GetResult } from "@fingerprintjs/fingerprintjs";

const fp = FingerprintJS.load();

export class DeviceInfo {
    /**
     * Get device fingerprint
     */
    async getFingerprint(): Promise<string | undefined> {
        try {
            const { visitorId } = await this.getInstance();

            return visitorId;
        } catch (error) {
            console.error("Failed to get device fingerprint", error);

            return undefined;
        }
    }

    /**
     * Get result instance
     */
    private async getInstance(): Promise<GetResult> {
        if (DeviceInfo.#device === undefined) {
            throw new Error("Device info not initialized");
        }

        const result = await DeviceInfo.#device;

        return result.get();
    }

    /**
     * Device info instance
     */
    static #device: Promise<Agent> | undefined;

    /**
     * Initialize device info. Should be called once on app start
     */
    static init(): void {
        DeviceInfo.#device = fp;
    }
}
