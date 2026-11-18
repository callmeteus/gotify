import * as React from 'react';
import axios from 'axios';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import * as config from '../config';
import DefaultPage from '../common/DefaultPage';
import {IPresence} from '../types';
import {presenceStateLabel} from './labels';

const Presence = () => {
    const [items, setItems] = React.useState<IPresence[]>([]);
    const [error, setError] = React.useState<string | null>(null);

    const load = React.useCallback(() => {
        axios
            .get<IPresence[]>(`${config.get('url')}presence`)
            .then((response) => {
                setItems(response.data);
                setError(null);
            })
            .catch(() => setError('Could not load presence'));
    }, []);

    React.useEffect(load, [load]);

    const override = (state: IPresence['state']) => {
        axios
            .put(`${config.get('url')}presence`, {
                state,
                source: 'manual',
                detail: null,
                validForSeconds: 1800,
            })
            .then(load)
            .catch(() => setError('Could not publish presence'));
    };

    return (
        <DefaultPage
            title="Presence"
            maxWidth={1000}
            rightControl={
                <div>
                    <Button onClick={() => override('at_desk')}>I'm at the PC</Button>
                    <Button onClick={() => override('away')}>I'm away</Button>
                </div>
            }>
            {error ? <Typography color="error">{error}</Typography> : null}
            <TableContainer id="presence-table" component={Paper}>
                <Table size="small">
                    <TableHead>
                        <TableRow>
                            <TableCell>User</TableCell>
                            <TableCell>State</TableCell>
                            <TableCell>Source</TableCell>
                            <TableCell>Detail</TableCell>
                            <TableCell>Reported</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {items.map((item) => (
                            <TableRow key={item.userId}>
                                <TableCell>{item.userId}</TableCell>
                                <TableCell>{presenceStateLabel(item.state)}</TableCell>
                                <TableCell>{item.source}</TableCell>
                                <TableCell>{item.detail ?? '-'}</TableCell>
                                <TableCell>{new Date(item.reportedAt).toLocaleString()}</TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </TableContainer>
        </DefaultPage>
    );
};

export default Presence;
