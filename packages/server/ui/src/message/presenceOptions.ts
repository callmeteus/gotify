import {IAlarmStage} from '../types';

export interface PresenceOption {
    label: string;
    state: 'at_desk' | 'away';
}

/** Builds the manual presence buttons from the alarm stages configured by the operator. */
export const presenceOptionsFromStages = (stages: IAlarmStage[]): PresenceOption[] =>
    stages.map((stage, index) => ({
        label: stage.label,
        state: index === stages.length - 1 ? 'away' : 'at_desk',
    }));
