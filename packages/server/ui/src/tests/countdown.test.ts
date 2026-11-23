import {describe, expect, it} from 'vitest';
import {countdownOf} from '../message/countdown';

const NOW = new Date('2026-01-01T00:00:00Z').getTime();

describe('countdownOf', () => {
    it('computes remaining time and fraction', () => {
        const result = countdownOf(new Date(NOW + 120_000).toISOString(), 600, NOW);
        expect(result?.remainingSeconds).toBe(120);
        expect(result?.fraction).toBeCloseTo(0.2);
        expect(result?.label).toBe('2:00');
    });

    it('clamps expired countdowns to zero', () => {
        const result = countdownOf(new Date(NOW - 1000).toISOString(), 600, NOW);
        expect(result?.remainingSeconds).toBe(0);
        expect(result?.fraction).toBe(0);
    });

    it('returns null without a deadline or duration', () => {
        expect(countdownOf(undefined, 600, NOW)).toBeNull();
        expect(countdownOf(new Date(NOW).toISOString(), 0, NOW)).toBeNull();
    });
});
