import {describe, expect, it} from 'vitest';
import {presenceOptionsFromStages} from '../message/presenceOptions';

describe('presenceOptionsFromStages', () => {
    it('uses only the labels configured in the alarm', () => {
        expect(
            presenceOptionsFromStages([
                {label: 'Estou ciente', advanceWithinSeconds: 600},
                {label: 'Estou no computador', advanceWithinSeconds: 0},
            ])
        ).toEqual([
            {label: 'Estou ciente', state: 'at_desk'},
            {label: 'Estou no computador', state: 'away'},
        ]);
    });

    it('returns no buttons without configured stages', () => {
        expect(presenceOptionsFromStages([])).toEqual([]);
    });
});
