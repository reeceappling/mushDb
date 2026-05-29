'use client'

import {HandleJsonResponse} from "@/app/components/common";
import {createContext, ReactNode, useContext, useEffect, useState} from "react";
import {CartesianGrid, Line, LineChart, Tooltip, XAxis, YAxis} from "recharts";
import {BaseExternalUrl} from "@/app/components/Constants";

interface TimespanContextInterface {
    timespan: Timespan
    setStartTime: (time: number) => void;
    setEndTime: (time?: number) => void;
}

interface Timespan {
    startTime: number; // EpochTimestampMs
    endTime?: number; // EpochTimestampMs
}

interface SensorData {
    T?: number
    H?: number
    L?: number
}

export const TimespanContext = createContext<TimespanContextInterface | undefined>(undefined);

export const useTimespanContext = () => {
    const context = useContext(TimespanContext);
    if (context === undefined) {
        throw new Error("useTimespanContext must be used within a TimespanProvider");
    }
    return context;
};

export interface DiscreteData {
    time: number; // EpochTimestampMs
    T?: number;
    H?: number;
    L?: number;
}

export function TimespanProvider(
    {
        props, children
    }: {
        props: {
            initialStartTimeMs: number
            initialEndTimeMs: number
        },
        children: ReactNode,
    }) {
    const [timespan, setTimespan] = useState<Timespan>({
        startTime: props.initialStartTimeMs,
        endTime: props.initialEndTimeMs
    });
    const setStartTime = (st: number) => {
        let upd = {...timespan}
        upd.startTime = st;
        setTimespan(upd)
    }
    const setEndTime = (et?: number) => {
        let upd = {...timespan}
        upd.endTime = et;
        setTimespan(upd)
    }
    return (
        <TimespanContext.Provider value={{timespan, setStartTime, setEndTime}}>
            <TimespanSelector/>
            {children}
        </TimespanContext.Provider>
    );
};

export function TimespanSelector() { // TODO: FIX
    const {timespan, setStartTime, setEndTime} = useTimespanContext()
    return <div>
        <div>{timespan.startTime}</div>
        {/* TODO: SELECTOR FOR START TIME*/}
        <div>{timespan.endTime}</div>
        {/* TODO: SELECTOR FOR START TIME*/}
    </div>
}


export function NodeMeasurementsViewer({nodeName}: { nodeName: string }) {
    const {timespan} = useTimespanContext()
    const [data, setData] = useState<DiscreteData[]>([])
    const [loaded, setLoaded] = useState(false)
    const [componentInterval, setComponentInterval] = useState(setInterval(() => {
    }, 10000))

    const fetchDataForTimespan: (span: Timespan, node: string)=>Promise<DiscreteData[]> = async (span: Timespan, node: string) => {
        return fetch(BaseExternalUrl+"/sensorData/"+node, { // TODO: create this endpoint
            method: 'GET',
            body: JSON.stringify(span), // TODO: ensure ok
            headers: {
                'Content-type': "application/json"
            },
        }).then(HandleJsonResponse)
            .then((results) => { // TODO: create this endpoint
                return results as DiscreteData[]
            })
    }
    const fetchNewData: (last: number, node: string)=>Promise<DiscreteData[]> = async (last: number, node: string) => {
        return fetch(BaseExternalUrl+"/sensorDataSince/"+node, {
            method: 'GET',
            body: JSON.stringify(last), // TODO: ensure ok
            headers: {
                'Content-type': "application/json"
            },
        }).then(HandleJsonResponse)
            .then((results) => {
                return results as DiscreteData[]
            })// TODO: fix
    }


    useEffect(()=>{
        const intv = setInterval(() => {
            if (timespan.endTime !== undefined) {
                 fetchNewData(data[data.length-1].time, nodeName).then((newData)=>{
                     if(newData.length>0){
                         setData([...data, ...newData])
                     }
                 }).catch(
                     (e)=> {
                         console.error("failed to fetch updated data: ",e)/* TODO: fix */
                     }
                 )

            }
        }, 1000)// Updates every 1000 milliseconds (1 second) // TODO: fix

        return () => {
            clearInterval(intv); // Cleanup on unmount
        };
    })
    useEffect(() => {
            fetchDataForTimespan(timespan, nodeName).then(setData).catch((e)=> {
                    console.error("failed to fetch span data: ",e)
                }/* TODO: fix */
            );
        },
        [timespan])

    return <LineChart
        width={500} // TODO: FIX
        height={300} // TODO: FIX
        data={data} // Data array (each object represents a point on the chart)
        margin={{
            top: 5,// TODO: FIX?
            right: 30,// TODO: FIX?
            left: 20,// TODO: FIX?
            bottom: 5,// TODO: FIX?
        }}
    >
        <CartesianGrid strokeDasharray={"3 3"/* Grid lines (dashed style) */}/>
        <XAxis
            dataKey="time"// TODO: FIX?
            padding={{left: 20, right: 20}}// TODO: FIX?
        />
        <YAxis/>{/* TODO: FIX? */}
        <Tooltip/>{/* TODO: FIX? */}

        <Line
            type={"monotone"}// TODO: FIX?
            dataKey={"T"}// TODO: FIX?
            stroke={"#8884d8"}// TODO: FIX?
            activeDot={{r: 8}}// TODO: FIX?
        />
        <Line
            type={"monotone"}// TODO: FIX?
            dataKey={"H"}// TODO: FIX?
            stroke={"#3334d8"}// TODO: FIX?
            activeDot={{r: 8}}// TODO: FIX?
        />
        <Line
            type={"monotone"}// TODO: FIX?
            dataKey={"L"}// TODO: FIX?
            stroke={"#CCC4d8"}// TODO: FIX?
            activeDot={{r: 8}}// TODO: FIX?
        />

    </LineChart>
}