'use client'

import React, {useState} from "react";
import WetnessSlider from "@/app/components/formSubcomponents/utils/slider";


export default function TestSlider({defaultValue}:{defaultValue?:number}){
    const [val, setVal] = useState(defaultValue || 5)
    const onChange = (event: Event, value: number, activeThumb: number) => {
        setVal(value)
    }
    return <div>
        <WetnessSlider onChange={onChange} defaultValue={defaultValue || 5}/>
        <div>{"current value: "+val}</div>
    </div>
}