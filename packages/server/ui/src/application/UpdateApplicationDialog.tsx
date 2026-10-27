import Add from '@mui/icons-material/Add';
import Delete from '@mui/icons-material/Delete';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import FormControlLabel from '@mui/material/FormControlLabel';
import IconButton from '@mui/material/IconButton';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import Tooltip from '@mui/material/Tooltip';
import {NumberField} from '../common/NumberField';
import {IAlarmConfig, IAlarmStage} from '../types';
import React, {useState} from 'react';

interface IProps {
    fClose: VoidFunction;
    fOnSubmit: (
        name: string,
        description: string,
        defaultPriority: number,
        alarm: IAlarmConfig
    ) => Promise<void>;
    initialName: string;
    initialDescription: string;
    initialDefaultPriority: number;
    initialAlarm?: IAlarmConfig;
}

const defaultStages = (): IAlarmStage[] => [
    {label: "I'm up", advanceWithinSeconds: 600},
    {label: "I'm at my computer", advanceWithinSeconds: 0},
];

const emptyAlarm = (): IAlarmConfig => ({
    enabled: false,
    priorityThreshold: 8,
    repeatIntervalSeconds: 30,
    maxRingSeconds: 1800,
    stages: defaultStages(),
});

/**
 * Tells whether an enabled alarm has two stages and a deadline on every stage except the last.
 *
 * @param alarm - Alarm the operator is about to save
 * @returns True when the alarm can be stored
 */
function alarmIsValid(alarm: IAlarmConfig): boolean {
    if (!alarm.enabled) {
        return true;
    }
    if (alarm.stages.length < 2) {
        return false;
    }
    if (alarm.repeatIntervalSeconds <= 0 || alarm.maxRingSeconds <= 0) {
        return false;
    }
    return alarm.stages.every((stage, index) => {
        if (stage.label.trim().length === 0) {
            return false;
        }
        const last = index === alarm.stages.length - 1;
        if (!last && stage.advanceWithinSeconds <= 0) {
            return false;
        }
        return true;
    });
}

/**
 * Edits an application's name, description, default priority, and alarm stages.
 *
 * @param props - Current application fields and the submit callback
 * @returns The update dialog
 */
export const UpdateApplicationDialog = ({
    initialName,
    initialDescription,
    initialDefaultPriority,
    initialAlarm,
    fClose,
    fOnSubmit,
}: IProps) => {
    const [name, setName] = useState(initialName);
    const [description, setDescription] = useState(initialDescription);
    const [defaultPriority, setDefaultPriority] = useState(initialDefaultPriority);
    const [alarm, setAlarm] = useState<IAlarmConfig>(initialAlarm ?? emptyAlarm());

    const submitEnabled = name.length !== 0 && alarmIsValid(alarm);
    let submitHint = '';
    if (name.length === 0) {
        submitHint = 'name is required';
    } else if (!alarmIsValid(alarm)) {
        submitHint = 'an enabled alarm needs two stages';
    }
    const submitAndClose = async () => {
        await fOnSubmit(name, description, defaultPriority, alarm);
        fClose();
    };
    const updateStage = (index: number, patch: Partial<IAlarmStage>) => {
        setAlarm({
            ...alarm,
            stages: alarm.stages.map((stage, stageIndex) => {
                if (stageIndex !== index) {
                    return stage;
                }
                return {
                    ...stage,
                    ...patch,
                };
            }),
        });
    };

    return (
        <Dialog
            open={true}
            onClose={fClose}
            aria-labelledby="form-dialog-title"
            id="app-dialog"
            maxWidth="sm"
            fullWidth>
            <DialogTitle id="form-dialog-title">Update an application</DialogTitle>
            <DialogContent>
                <DialogContentText>An application is allowed to send messages.</DialogContentText>
                <TextField
                    autoFocus
                    margin="dense"
                    className="name"
                    label="Name *"
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    fullWidth
                />
                <TextField
                    margin="dense"
                    className="description"
                    label="Short Description"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    fullWidth
                    multiline
                />
                <NumberField
                    margin="dense"
                    className="priority"
                    label="Default Priority"
                    value={defaultPriority}
                    onChange={(e) => setDefaultPriority(e)}
                    fullWidth
                />
                <FormControlLabel
                    control={
                        <Switch
                            className="alarm-enabled"
                            checked={alarm.enabled}
                            onChange={(event) =>
                                setAlarm({
                                    ...alarm,
                                    enabled: event.target.checked,
                                })
                            }
                        />
                    }
                    label="Alarm"
                />
                <NumberField
                    margin="dense"
                    className="alarm-threshold"
                    label="Priority threshold"
                    value={alarm.priorityThreshold}
                    onChange={(value) =>
                        setAlarm({
                            ...alarm,
                            priorityThreshold: value,
                        })
                    }
                    fullWidth
                />
                <NumberField
                    margin="dense"
                    className="alarm-repeat"
                    label="Repeat interval (seconds)"
                    value={alarm.repeatIntervalSeconds}
                    onChange={(value) =>
                        setAlarm({
                            ...alarm,
                            repeatIntervalSeconds: value,
                        })
                    }
                    fullWidth
                />
                <NumberField
                    margin="dense"
                    className="alarm-max-ring"
                    label="Max ringing (seconds)"
                    value={alarm.maxRingSeconds}
                    onChange={(value) =>
                        setAlarm({
                            ...alarm,
                            maxRingSeconds: value,
                        })
                    }
                    fullWidth
                />
                {alarm.stages.map((stage, index) => (
                    <div key={index}>
                        <TextField
                            margin="dense"
                            className="alarm-stage-label"
                            label={`Stage ${index + 1}`}
                            value={stage.label}
                            onChange={(event) =>
                                updateStage(index, {
                                    label: event.target.value,
                                })
                            }
                            fullWidth
                        />
                        <NumberField
                            margin="dense"
                            className="alarm-stage-seconds"
                            label="Advance within (seconds)"
                            value={stage.advanceWithinSeconds}
                            onChange={(value) =>
                                updateStage(index, {
                                    advanceWithinSeconds: value,
                                })
                            }
                            fullWidth
                        />
                        <IconButton
                            aria-label="Remove stage"
                            onClick={() =>
                                setAlarm({
                                    ...alarm,
                                    stages: alarm.stages.filter((_, stageIndex) => {
                                        return stageIndex !== index;
                                    }),
                                })
                            }>
                            <Delete />
                        </IconButton>
                    </div>
                ))}
                <Button
                    startIcon={<Add />}
                    onClick={() =>
                        setAlarm({
                            ...alarm,
                            stages: alarm.stages.concat({
                                label: '',
                                advanceWithinSeconds: 600,
                            }),
                        })
                    }>
                    Add stage
                </Button>
            </DialogContent>
            <DialogActions>
                <Button onClick={fClose}>Cancel</Button>
                <Tooltip title={submitHint}>
                    <div>
                        <Button
                            className="update"
                            disabled={!submitEnabled}
                            onClick={submitAndClose}
                            color="primary"
                            variant="contained">
                            Update
                        </Button>
                    </div>
                </Tooltip>
            </DialogActions>
        </Dialog>
    );
};
