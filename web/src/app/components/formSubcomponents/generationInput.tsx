import {useState} from "react";
import {InputNumber} from "@/app/components/formSubcomponents/numericInput";

export function GenerationInput({
                                           updateParent,
                                       }: {
    updateParent: (gen: number | undefined) => void,
}) {
    const [current, setCurrent] = useState<number | undefined>(1)
    return (
        <div className={"inlineChildren"}>
            <div className={"text-lg mr-2"}>{"Generation: "}</div>
            <InputNumber min={0} max={1000} step={1} value={(current || 0).toString()} readonly={false} mode={"integer"} placeholder={"gen"} onChange={(v) => {
                try {
                    const val = Number(v)
                    if (val===0){
                        updateParent(undefined) // TODO: validate ok
                        setCurrent(undefined) // TODO: validate ok
                    } else {
                        updateParent(val)
                        setCurrent(val)
                    }
                } catch (e){
                    // TODO: unsure what to do here
                }
            }}/>
            <div className={"text-md ml-2"}>{"(Put 0 for unknown)"}</div>
        </div>
    )
}