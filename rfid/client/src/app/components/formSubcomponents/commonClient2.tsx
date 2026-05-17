import {NewPicWithNotesForm, PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {useEffect, useState} from "react";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import DateArea from "@/app/components/formSubcomponents/date";
import {Note} from "@/app/components/formSubcomponents/notes";
import {AllEntries, Data} from "@/app/components/formSubcomponents/shared";
import {NotesFormArea} from "@/app/components/agarBatchClient";

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
                let upd = structuredClone(current)
                upd.img = f
                updateRow(upd)
            }}/>
            <button className={"removeButton"} onClick={remv}>{"REMOVE THIS Entry"}</button>
        </div>
    }
    const rightArea = () => {
        return <div className={"picRight"}>
            <DateArea readonly={true} when={current.time}/>
            <NotesFormArea readonly={false} initial={[]} updateParent={(nts: AllEntries<Note>) => {
                let updated = structuredClone(current)
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

export function PixRowsNew(
    {initial, updateParent}: {
        initial: PicWithNotesIncoming[],
        updateParent?: (d: NewPicWithNotesForm[]) => void,
    }) {
    const [current, setCurrent] = useState<Data<NewPicWithNotesForm>[]>([])
    useEffect(() => {
        setCurrent([])  // Reset when initial changes
    }, [initial])
    const doUpdate = (updated: Data<NewPicWithNotesForm>[]) => {
        setCurrent(updated)
        updateParent && updateParent(updated.filter(e => {
            const hasImgOrNotes = e.data.img !== undefined || e.data.notes.new.length > 0
            return !e.disabled && hasImgOrNotes
        }).map(e => e.data))
    }
    return <>
        <div className={"picsGroup picsRows"}>
            {current.map((v, i) => {
                return <PixRowNew remv={() => {
                    let upd = structuredClone(current)
                    upd[i].disabled = true
                    doUpdate(upd)
                }} updateParent={(u) => {
                    let upd = structuredClone(current)
                    upd[i].data = u
                    doUpdate(upd)
                }}/>
            })}
        </div>
        <div className={"centerH gapTop"}>
            <button className={"greenButton"} onClick={(e) => {
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
            }}>{"Add picture"}</button>
        </div>
    </>

}