import {IPresence} from '../types';

/** Labels shown for the presence states in the UI. */
export const PRESENCE_STATE_LABELS: Record<IPresence['state'], string> = {
    at_desk: 'At the PC',
    away: 'Away',
    unknown: 'Unknown',
};

/** Returns the user-facing label, falling back to the wire value for unknown states. */
export function presenceStateLabel(state: IPresence['state']): string {
    return Object.prototype.hasOwnProperty.call(PRESENCE_STATE_LABELS, state)
        ? PRESENCE_STATE_LABELS[state]
        : state;
}
