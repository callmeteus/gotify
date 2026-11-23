import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import LinearProgress from '@mui/material/LinearProgress';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import axios from 'axios';
import React from 'react';
import * as config from '../config';
import {countdownOf} from './countdown';
import {IIncident} from '../types';

const OPTIONS = [
    {label: "I'm up", state: 'at_desk'},
    {label: "I'm at my computer", state: 'at_desk'},
    {label: 'Not at PC', state: 'away'},
] as const;

/** Manual presence controls and the active alarm countdown, always visible. */
const PresenceControls = ({incident}: {incident?: IIncident}) => {
    const [current, setCurrent] = React.useState<string | null>(null);
    const [now, setNow] = React.useState(Date.now());

    React.useEffect(() => {
        const timer = window.setInterval(() => setNow(Date.now()), 1000);
        return () => window.clearInterval(timer);
    }, []);

    const publish = async (state: string, label: string) => {
        try {
            await axios.put(`${config.get('url')}presence`, {
                state,
                source: 'manual',
                detail: label,
                validForSeconds: 1800,
            });
            setCurrent(label);
        } catch {
            setCurrent(null);
        }
    };

    const countdown =
        incident && (incident.state === 'ringing' || incident.state === 'waiting')
            ? countdownOf(incident.deadlineAt, incident.durationSeconds, now)
            : null;

    return (
        <Paper id="presence-controls" elevation={6} style={{padding: 12, marginBottom: 16}}>
            <Typography variant="subtitle2" style={{marginBottom: 8}}>
                Presence {current ? `· ${current}` : ''}
            </Typography>
            <div style={{display: 'flex', gap: 8, flexWrap: 'wrap'}}>
                {OPTIONS.map((option) => (
                    <Button
                        key={option.label}
                        size="small"
                        variant={current === option.label ? 'contained' : 'outlined'}
                        onClick={() => publish(option.state, option.label)}>
                        {option.label}
                    </Button>
                ))}
            </div>
            {countdown && (
                <Box id="incident-countdown" sx={{width: '100%', mt: 2}}>
                    <Box sx={{display: 'flex', justifyContent: 'space-between', mb: 0.5}}>
                        <Typography variant="caption">
                            {incident?.stages[incident.currentStage]?.label ?? 'Alarm'}
                        </Typography>
                        <Typography variant="caption">{countdown.label}</Typography>
                    </Box>
                    <LinearProgress
                        variant="determinate"
                        value={countdown.fraction * 100}
                        sx={{width: '100%', height: 8, borderRadius: 4}}
                    />
                </Box>
            )}
        </Paper>
    );
};

export default PresenceControls;
