import {describe, expect, it} from 'vitest';
import {presenceStateLabel} from '../presence/labels';

describe('presenceStateLabel', () => {
    it('labels at_desk', () => expect(presenceStateLabel('at_desk')).toBe('At the PC'));
    it('labels away', () => expect(presenceStateLabel('away')).toBe('Away'));
    it('labels unknown', () => expect(presenceStateLabel('unknown')).toBe('Unknown'));
    it('falls back to the wire value', () =>
        expect(presenceStateLabel('bogus' as 'at_desk')).toBe('bogus'));
    it('does not read prototype keys as labels', () =>
        expect(presenceStateLabel('toString' as 'at_desk')).toBe('toString'));
});
