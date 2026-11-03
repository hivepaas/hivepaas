interface QueueItem {
    onSuccess: (token: string) => void;
    onFailed: () => void;
}

export class RefreshQueue {
    private isQueueOpen = false;
    private queue: QueueItem[] = [];

    open(): void {
        this.isQueueOpen = true;
    }

    add(item: QueueItem): void {
        this.queue.push(item);
    }

    success(token: string): void {
        this.isQueueOpen = false;

        this.queue.forEach(item => {
            item.onSuccess(token);
        });

        this.queue = [];
    }

    failed(): void {
        this.isQueueOpen = false;

        this.queue.forEach(item => {
            item.onFailed();
        });

        this.queue = [];
    }

    get isOpen(): boolean {
        return this.isQueueOpen;
    }
}
