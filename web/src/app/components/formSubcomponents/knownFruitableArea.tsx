import {SyntheticEvent, useState} from "react";

export function KnownFruitableArea(
    {
        initial,
        readonly,
        doSelect,
    }: {
        initial?: boolean
        readonly?: boolean
        doSelect?: (val: boolean | undefined) => void
        headerLevel?: number
    }) {
    const KFTxt = "Known Fruitable: "
    if (readonly) { // TODO: ENSURE OK
        return <div className={"knownFruitableArea"}>
            {KFTxt + ": " + optionalBoolToString(initial)}
        </div>
    }
    const [knownFruitable, setKnownFruitable] = useState<string | undefined>(optionalBoolToOptionalString(initial))
    const onSelect = (e: SyntheticEvent<HTMLSelectElement, Event>) => {
        let val = e.currentTarget.value
        let selectFunc = (doSelect === undefined) ? () => {
        } : doSelect
        selectFunc((val === "" || val === "unknown") ? undefined : val === "true")
        setKnownFruitable(val)
    }
    return <div className={"knownFruitableArea"}>
        <div>{KFTxt}</div>
        <div>
            <select className={"tailwindSelector ml-1"} value={knownFruitable || ""} onChange={onSelect}>
                {["", "true", "false", "unknown"].map((val, i) => {
                    return <option value={val} key={i}>{val}</option>
                })}
            </select>
            {/* TODO: RESET BUTTON? */}
        </div>
    </div>
}

function optionalBoolToOptionalString(b?: boolean) {
    if (b == undefined) {
        return b
    }
    return b ? "true" : "false"
}

function optionalBoolToString(b?: boolean) {
    if (b == undefined) {
        return "unknown"
    }
    return b ? "true" : "false"
}