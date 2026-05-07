import {useState} from "react";
import {NumberStringOnlyFromText} from "@/app/components/formSubcomponents/date";
import {InputNumberWithSmallTitle} from "@/app/components/formSubcomponents/numericInput";

export default function GenerationArea({
                                           initial,
                                           labelName,
                                           readonly,
                                           updateParent,
                                           headerLevel
                                       }: {
    initial?: number,
    readonly?: boolean,
    labelName?: string,
    updateParent: (gen: number | undefined) => void,
    headerLevel?: number
}) {
    const [current, setCurrent] = useState<number | undefined>(initial)
    const GenMarker = (labelName || "Generation") + ": "
    if (readonly) {
        if (current == undefined) {
            return <div className={"areaHeader"}>{GenMarker + "unknown"}</div>
        }
        return <div>{GenMarker + String(current)}</div>
    }
    return (
        <div className={"gapBottom"}>
            <div className={"areaHeader"}>{GenMarker}</div>
            <div className={"centerH"}>
                {/* TODO: ensure ok*/}
                <InputNumberWithSmallTitle label={"Generation"} min={0} max={1000} step={1} value={(current || 0).toString()}
                                           readonly={false} mode={"integer"} placeholder={"gen"} onChange={(v) => {
                    try {
                        let val = Number(v)
                        updateParent(val)
                        setCurrent(val)
                    } catch (e) {
                        console.error(e)
                    }
                }}/>
            </div>
        </div>
    )
}