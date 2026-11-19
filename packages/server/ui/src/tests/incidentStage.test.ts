import {describe, expect, it} from 'vitest';
import {nextIncidentStage} from '../message/incident';
import {IIncident} from '../types';

const incident = (state: string, currentStage: number): IIncident => ({
    id: 1,
    appid: 2,
    state,
    kind: '',
    currentStage,
    stages: [
        {label: "I'm up", advanceWithinSeconds: 600},
        {label: "I'm at my computer", advanceWithinSeconds: 0},
    ],
    repeatIntervalSeconds: 30,
    title: '',
    message: '',
});

describe('nextIncidentStage', () => {
    it('starts at stage zero while ringing', () =>
        expect(nextIncidentStage(incident('ringing', -1))).toBe(0));
    it('moves to the following stage while waiting', () =>
        expect(nextIncidentStage(incident('waiting', 0))).toBe(1));
    it('returns null when closed', () =>
        expect(nextIncidentStage(incident('closed', 1))).toBeNull());
    it('returns null when no incident is attached', () =>
        expect(nextIncidentStage(undefined)).toBeNull());
    it('returns null when the stage index is out of range', () =>
        expect(nextIncidentStage(incident('waiting', 1))).toBeNull());
});
