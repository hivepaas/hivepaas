interface Message {
    bitmap: ImageBitmap;
    filename: string;
    quality: number;
    type: string;
}

self.onmessage = async (event: MessageEvent<Message>) => {
    const { bitmap, quality, type, filename } = event.data;

    const offscreen = new OffscreenCanvas(bitmap.width, bitmap.height);
    const ctx = offscreen.getContext("2d");

    if (!ctx) {
        self.postMessage({ error: "Could not create offscreen canvas context" });

        return;
    }

    ctx.drawImage(bitmap, 0, 0);

    try {
        const blob = await offscreen.convertToBlob({ type, quality });

        const file = new File([blob], filename, {
            type,
            lastModified: Date.now(),
        });

        self.postMessage(file);
    } catch (error) {
        self.postMessage({ error });
    }
};
