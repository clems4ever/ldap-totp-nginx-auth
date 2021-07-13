import React, { ReactNode, useState } from "react";

import { makeStyles, Typography, Grid, Button, Container } from "@material-ui/core";

import PushNotificationIcon from "@components/PushNotificationIcon";

export enum State {
    DEVICE = 1,
    METHOD = 2,
}

export interface SelectableDevice {
    id: string;
    name: string;
    methods: string[];
}

export interface SelectedDevice {
    id: string;
    method: string;
}

export interface Props {
    children?: ReactNode;
    devices: SelectableDevice[];

    onBack: () => void;
    onSelect: (device: SelectedDevice) => void;
}
const DefaultDeviceSelectionContainer = function (props: Props) {
    const [state, setState] = useState(State.DEVICE);
    // eslint-disable-next-line prettier/prettier
    const [device, setDevice] = useState([] as unknown as SelectableDevice);

    const handleDeviceSelected = (selecteddevice: SelectableDevice) => {
        if (selecteddevice.methods.length === 1) handleMethodSelected(selecteddevice.methods[0], selecteddevice.id);
        else {
            setDevice(selecteddevice);
            setState(State.METHOD);
        }
    };

    const handleMethodSelected = (method: string, deviceid?: string) => {
        if (deviceid) props.onSelect({ id: deviceid, method: method });
        else props.onSelect({ id: device.id, method: method });
    };

    let container: ReactNode;
    switch (state) {
        case State.DEVICE:
            container = (
                <Grid container justify="center" spacing={1} id="device-selection">
                    {props.devices.map((value, index) => {
                        return (
                            <DeviceItem
                                id={index}
                                key={index}
                                device={value}
                                onSelect={() => handleDeviceSelected(value)}
                            />
                        );
                    })}
                </Grid>
            );
            break;
        // TODO
        case State.METHOD:
            container = (
                <Grid container justify="center" spacing={1} id="method-selection">
                    {device.methods.map((value, index) => {
                        return (
                            <MethodItem
                                id={index}
                                key={index}
                                method={value}
                                onSelect={() => handleMethodSelected(value)}
                            />
                        );
                    })}
                </Grid>
            );
            break;
    }

    return (
        <Container>
            {container}
            <Button color="primary" onClick={props.onBack}>
                back
            </Button>
        </Container>
    );
};

export default DefaultDeviceSelectionContainer;

interface DeviceItemProps {
    id: number;
    device: SelectableDevice;

    onSelect: () => void;
}

function DeviceItem(props: DeviceItemProps) {
    const className = "device-option-" + props.id;
    const style = makeStyles((theme) => ({
        item: {
            paddingTop: theme.spacing(4),
            paddingBottom: theme.spacing(4),
            width: "100%",
        },
        icon: {
            display: "inline-block",
            fill: "white",
        },
        buttonRoot: {
            display: "block",
        },
    }))();

    return (
        <Grid item xs={12} className={className} id={props.device.id}>
            <Button
                className={style.item}
                color="primary"
                classes={{ root: style.buttonRoot }}
                variant="contained"
                onClick={props.onSelect}
            >
                <div className={style.icon}>
                    <PushNotificationIcon width={32} height={32} />
                </div>
                <div>
                    <Typography>{props.device.name}</Typography>
                </div>
            </Button>
        </Grid>
    );
}

interface MethodItemProps {
    id: number;
    method: string;

    onSelect: () => void;
}

function MethodItem(props: MethodItemProps) {
    const className = "method-option-" + props.id;
    const style = makeStyles((theme) => ({
        item: {
            paddingTop: theme.spacing(4),
            paddingBottom: theme.spacing(4),
            width: "100%",
        },
        icon: {
            display: "inline-block",
            fill: "white",
        },
        buttonRoot: {
            display: "block",
        },
    }))();

    return (
        <Grid item xs={12} className={className} id={props.method}>
            <Button
                className={style.item}
                color="primary"
                classes={{ root: style.buttonRoot }}
                variant="contained"
                onClick={props.onSelect}
            >
                <div className={style.icon}>
                    <PushNotificationIcon width={32} height={32} />
                </div>
                <div>
                    <Typography>{props.method}</Typography>
                </div>
            </Button>
        </Grid>
    );
}
