'use client'
import {SyntheticEvent, useState} from "react";
import {NumericalArea} from "@/app/components/formSubcomponents/numericInput";

function VolumeSelectorInternal({options,initialValue,onChange}:
                                      {
                                          options:string[],
                                          initialValue:string,
                                          onChange: (value: string) => void,
                                      }) {
    const [current, setCurrent] = useState(initialValue)
    const onSelect = (e: SyntheticEvent<HTMLSelectElement, Event>) => {
        setCurrent(e.currentTarget.value)
        onChange(e.currentTarget.value)
    }
    return <select className="tailwindSelector" value={current} onChange={onSelect}>
        {options.map(function (name, i) {
            return <option value={name} key={i}>{name}</option>
        })}
    </select>
}

export function JarSizeSelector({onChange}:{onChange: (value: string)=>void}){
    return <VolumeSelectorInternal options={["pint","quart"]} initialValue={"quart"} onChange={onChange}/>
}

export function VolumeSelector({onChange,defaultUnit}:{defaultUnit?:string,onChange: (nCups: number)=>void}){
    const [num, setNum] = useState(1)
    const [vol, setVol] = useState(defaultUnit || "quarts")
    const handleChange = (n: number, s: string)=>{
        let mul = 1
        if(s==="pints"){
            mul = 2
        } else if(s==="quarts"){
            mul = 4
        }
        onChange(num * mul)
    }
    return <>
        <NumericalArea min={0.0} max={50.0} mode={'floating'} onChange={(s?: string)=>{
            if(!s){
                return
            }
            const newNum = Number(s)
            handleChange(newNum, vol)
            setNum(newNum)
        }}/>{/* TODO: ENSURE WORKS */}
        <VolumeSelectorInternal options={["cups","pints","quarts"]} initialValue={defaultUnit || "quarts"} onChange={(s: string)=>{
            handleChange(num, s)
            setVol(s)
        }}/>
    </>
}