import {useState} from "react";
import {NumberStringOnlyFromText} from "@/app/components/formSubcomponents/date";
import H from "@/app/components/formSubcomponents/utils/headers";

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
    const GenMarker = (labelName || "Generation")+": "
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
                <input type="text" name="generationInput" value={current || ""} onChange={(e) => {
                let str = NumberStringOnlyFromText(e.currentTarget.value)
                let val = (str === "") ? undefined : Number(str)
                updateParent(val)
                setCurrent(val)
            }}/>
            </div>
        </div>
    )
}