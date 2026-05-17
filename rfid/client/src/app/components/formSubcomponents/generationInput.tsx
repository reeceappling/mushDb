import {useState} from "react";
import {NumberStringOnlyFromText} from "@/app/components/formSubcomponents/date";
import {InputNumber, InputNumberWithSmallTitle} from "@/app/components/formSubcomponents/numericInput";
import {InputTextWithInlineTitle} from "@/app/components/formSubcomponents/commonClient";

// export default function GenerationArea({
//                                            initial,
//                                            labelName,
//                                            readonly,
//                                            updateParent,
//                                            headerLevel
//                                        }: {
//     initial?: number,
//     readonly?: boolean,
//     labelName?: string,
//     updateParent: (gen: number | undefined) => void,
//     headerLevel?: number
// }) {
//     const [current, setCurrent] = useState<number | undefined>(initial)
//     const GenMarker = (labelName || "Generation") + ": "
//     if (readonly) {
//         if (current == undefined) {
//             return <div className={"areaHeader"}>{GenMarker + "unknown"}</div>
//         }
//         return <div>{GenMarker + String(current)}</div>
//     }
//     return (
//         <div className={"gapBottom"}>
//             <div className={"areaHeader"}>{GenMarker}</div>
//             <div className={"centerH"}>
//                 {/* TODO: ensure ok*/}
//                 <InputNumberWithSmallTitle label={"Generation"} min={0} max={1000} step={1} value={(current || 0).toString()}
//                                            readonly={false} mode={"integer"} placeholder={"gen"} onChange={(v) => {
//                     try {
//                         let val = Number(v)
//                         updateParent(val)
//                         setCurrent(val)
//                     } catch (e) {
//                         console.error(e)
//                     }
//                 }}/>
//             </div>
//         </div>
//     )
// }
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
                    let val = Number(v)
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