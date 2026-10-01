import {NewPicWithNotesForm, PicWithNotesIncoming} from "@/app/components/formSubcomponents/picWithNotes";
import {useEffect, useState} from "react";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import DateArea from "@/app/components/formSubcomponents/date";
import {Note, NotesFormAreaPics} from "@/app/components/formSubcomponents/notes";
import {AllEntries, Data} from "@/app/components/formSubcomponents/shared";

function picRowsKey(items: PicWithNotesIncoming[]): string {
    return items.map((p) =>
        [
            p.time,
            p.location || "",
            (p.notes || []).map((n) => `${n.time}:${n.note}`).join("^"),
        ].join("|")
    ).join("||");
}

export function PixRows( // For Rows of new pictures (not preexisting)
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
                return <PixRowNew key={i} remv={() => {
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
            <NotesFormAreaPics readonly={false} initial={[]} allowLargeTextBox={false/* TODO: OK?*/}
                updateParent={(nts: AllEntries<Note>) => {
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
