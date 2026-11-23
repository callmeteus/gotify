export interface Countdown {
    remainingSeconds: number;
    fraction: number;
    label: string;
}

const format = (seconds: number) => {
    const whole = Math.max(0, Math.ceil(seconds));
    const minutes = Math.floor(whole / 60);
    const rest = whole % 60;
    return `${minutes}:${String(rest).padStart(2, '0')}`;
};

/** Remaining time of an incident countdown, independent of alarm internals. */
export const countdownOf = (
    deadlineAt?: string,
    durationSeconds?: number,
    now: number = Date.now()
): Countdown | null => {
    if (!deadlineAt || !durationSeconds || durationSeconds <= 0) {
        return null;
    }
    const remainingMs = new Date(deadlineAt).getTime() - now;
    if (!Number.isFinite(remainingMs)) {
        return null;
    }
    const remainingSeconds = Math.max(0, remainingMs / 1000);
    return {
        remainingSeconds,
        fraction: Math.max(0, Math.min(1, remainingSeconds / durationSeconds)),
        label: format(remainingSeconds),
    };
};
