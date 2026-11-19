import {IIncident} from '../types';

/** The stage index the incident is waiting for, or null when nothing can be confirmed. */
export const nextIncidentStage = (incident?: IIncident): number | null => {
    if (!incident || (incident.state !== 'ringing' && incident.state !== 'waiting')) {
        return null;
    }
    const next = incident.state === 'ringing' ? 0 : incident.currentStage + 1;
    return next >= 0 && next < incident.stages.length ? next : null;
};
