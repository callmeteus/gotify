import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import axios from 'axios';
import React from 'react';
import * as config from '../config';

const OPTIONS = [
    {label: "I'm up", state: 'at_desk'},
    {label: "I'm at my computer", state: 'at_desk'},
    {label: 'Not at PC', state: 'away'},
] as const;

/** Manual presence controls, available on every message page like in the Android app. */
const PresenceControls = () => {
    const [current, setCurrent] = React.useState<string | null>(null);

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
        </Paper>
    );
};

export default PresenceControls;
