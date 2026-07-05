import {NewPicWithNotesForm, PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {useEffect, useState} from "react";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import DateArea from "@/app/components/formSubcomponents/date";
import {Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AllEntries, Data} from "@/app/components/formSubcomponents/shared";
import {InputDecimal, InputNumber, InputText} from "@/app/components/formSubcomponents/numericInput";

function picRowsKey(items: PicWithNotesIncoming[]): string {
    return items.map((p) =>
        [
            p.time,
            p.location || "",
            (p.notes || []).map((n) => `${n.time}:${n.note}`).join("^"),
        ].join("|")
    ).join("||");
}

export function PixRows(
    {initial, updateParent, addButtonText}: {
        initial: PicWithNotesIncoming[],
        updateParent?: (d: NewPicWithNotesForm[]) => void,
        addButtonText?: string,
    }) {
    const [current, setCurrent] = useState<Data<NewPicWithNotesForm>[]>([])
    const initialKey = picRowsKey(initial);

    useEffect(() => {
        setCurrent([]); // Reset only when pic content actually changes
    }, [initialKey]);
    const doUpdate = (updated: Data<NewPicWithNotesForm>[]) => {
        setCurrent(updated)
        updateParent && updateParent(updated.filter(e => {
            const hasImgOrNotes = (e.data.img !== undefined) || (e.data.notes.new.length > 0)
            return !e.disabled && hasImgOrNotes
        }).map(e => e.data))
    }
    return <>
        <div className={"picsGroup picsRows"}>
            {current.map((v, i) => {
                return <PixRowNew key={i} remv={() => { // TODO: ensure key ok
                    const upd = structuredClone(current)
                    upd[i].disabled = true
                    doUpdate(upd)
                }} updateParent={(u) => {
                    const upd = structuredClone(current)
                    upd[i].data = u
                    doUpdate(upd)
                }}/>
            })}
        </div>
        <div className={"centerH gapTop picsRowsAdd"}>
            <button className={"greenButton"} onClick={(e) => {
                console.log("adding a picture")
                e.preventDefault();
                e.stopPropagation();
                const upd = [...structuredClone(current), {
                    data: {
                        time: Date.now(),
                        img: undefined,
                        notes: {existing: [], new: []}
                    },
                    disabled: false
                }]
                doUpdate(upd)
            }}>{addButtonText || "Add picture"}</button>
        </div>
    </>
}

export function PixRowNew(
    {updateParent, remv}: {
        updateParent?: (d: NewPicWithNotesForm) => void,
        remv: () => void,
    }) {
    const [current, setCurrent] = useState<NewPicWithNotesForm>({
        time: Date.now(),
        img: undefined,
        notes: {existing: [], new: []}
    });
    const updateRow = (updated: NewPicWithNotesForm) => {
        setCurrent(updated)
        updateParent && updateParent(updated)
    }
    const leftArea = () => {
        return <div className={"picLeft"}>
            {/* TODO: IMAGE AREA GROW/SHRINK ON CLICK */}
            <ImageSelector updateParent={f => {
                const upd = structuredClone(current)
                upd.img = f
                updateRow(upd)
            }}/>
            <button className={"removeButton"} onClick={remv}>{"Remove This Entry"}</button>
        </div>
    }
    const rightArea = () => {
        return <div className={"picRight"}>
            <DateArea readonly={true} when={current.time}/>
            <NotesFormArea readonly={false} initial={[]} updateParent={(nts: AllEntries<Note>) => {
                const updated = structuredClone(current)
                updated.notes = nts
                updateRow(updated)
            }} removeHeader={true}/>
        </div>
    }
    return <div className={"contentsOnly picRow"}>
        {leftArea()}
        {rightArea()}
    </div>
}

export function VolumetricInput(
    {
        initialParentVolmL,
        initialVolumetricAmount,
        initialVolumetricUnits,
    }:{
        initialParentVolmL?: number,
        initialVolumetricAmount?: number
        initialVolumetricUnits?: string
    }){
    const [parentVolmL, setParentVolmL] = useState(initialParentVolmL || 1000.0)
    const [volumetricValue, setVolumetricValue] = useState<number>(initialVolumetricAmount||0.0)
    const resolveChildVal = (volumetricMmt:number, parentmL?:number)=> {
        if (parentmL === undefined){
            return volumetricMmt
        }
        return (parentmL*volumetricMmt)/1000.0
    }
    const [childAmt, setChildAmt] = useState<number>(resolveChildVal(volumetricValue,parentVolmL))
    const [err, setErr] = useState<string | undefined>(undefined)
    const getInitialSubUnit = ()=>{
        if (!initialVolumetricUnits){
            return ""
        }
        const subUnits = initialVolumetricUnits.split("/")
        if (subUnits.length !== 2) {
            throw "invalid input units, must be 'x/y' where x and y are units of measure"
        } else {
            return subUnits[0]
        }
    }
    const getInitialSubUnitStart = ()=>{
        try {
            return getInitialSubUnit()
        } catch {
            setErr("invalid input units, must be 'x/y' where x and y are units of measure")
            return ""
        }
    }
    const [subUnit, setSubUnit] = useState<string>(getInitialSubUnitStart())
    // Handle changes to initial parent volume
    useEffect(()=>{
        // TODO: set child amount!
        if (initialParentVolmL === undefined){
            // TODO: what to do if no longer defined? no change?
            return
        }
        setParentVolmL(initialParentVolmL)
        setChildAmt(resolveChildVal(volumetricValue, initialParentVolmL))
    },[initialParentVolmL])
    // Handle changes to initial valumetric amount
    useEffect(()=>{
        // TODO: set child amount!
        if (initialVolumetricAmount===undefined){
            // TODO: ??? nothing?
            return
        }
        setVolumetricValue(initialVolumetricAmount)
        setChildAmt(resolveChildVal(initialVolumetricAmount, initialParentVolmL))
    },[initialVolumetricAmount])
    // Handle input units changes
    useEffect(()=>{
        try {
            setSubUnit(getInitialSubUnit())
        } catch{
            setErr("invalid input units, must be 'x/y' where x and y are units of measure")
        }
    },[initialVolumetricUnits])
    const actualValueInput = <InputDecimal initial={childAmt} min={0} label={"fixme smaller"} updateParent={(n:number)=>{
        setChildAmt(n)
        setVolumetricValue((n*1000)/parentVolmL)
    } }/>
    // const mlInput = <InputNumber value={mL.toString()} onChange={s=>{
    //     try {
    //         s.
    //     } catch (e){
    //
    //     }
    // }} updateParent={(n:number)=>{}} min={0} /*onChange={()=>{}}*/ />
    const subUnitInput = <InputText readonly={false} value={subUnit} errorMessage={undefined/*TODO: ???*/} onBlur={()=>{/*TODO: ???*/}} placeholder={undefined/*TODO: ???*/}/>
    const finalValueInput = <InputDecimal initial={volumetricValue} min={0} label={"fixme final"} updateParent={(n:number)=>{
        setVolumetricValue(n)
        setChildAmt(resolveChildVal(n, parentVolmL))
    } }/>
    return <div className={"inlineChildren"}>
        <div>{actualValueInput}</div>
        <div>{subUnitInput}</div>
        <div>{"("}{finalValueInput}{" "+subUnit+")"}</div>
    </div>
}